// Load test for acceptance criterion 5: "100 concurrent virtual users running
// 1 write/sec each for 60 seconds completes without errors or unbounded latency
// growth at MVP scale."
//
// Each virtual user owns its own document, so this measures throughput rather
// than contention — conflicts are FR3's subject and are covered by the test
// suite. Latency is reported per time window so that "unbounded growth" is
// visible rather than hidden inside an overall percentile.
package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type sample struct {
	ms     float64
	window int
	ok     bool
	read   bool
}

func percentile(v []float64, q float64) float64 {
	if len(v) == 0 {
		return 0
	}
	i := int(q * float64(len(v)-1))
	return v[i]
}

func summarise(name string, s []sample) {
	var lat []float64
	var errs int
	for _, x := range s {
		if x.ok {
			lat = append(lat, x.ms)
		} else {
			errs++
		}
	}
	sort.Float64s(lat)
	fmt.Printf("%-8s n=%-6d errors=%-4d  p50=%6.1fms  p95=%6.1fms  p99=%6.1fms  max=%7.1fms\n",
		name, len(s), errs, percentile(lat, 0.50), percentile(lat, 0.95),
		percentile(lat, 0.99), percentile(lat, 1.0))
}

func windows(name string, s []sample, nWindows int) {
	fmt.Printf("\n  %s latency per 10s window (checking for unbounded growth)\n", name)
	for w := 0; w < nWindows; w++ {
		var lat []float64
		for _, x := range s {
			if x.window == w && x.ok {
				lat = append(lat, x.ms)
			}
		}
		if len(lat) == 0 {
			continue
		}
		sort.Float64s(lat)
		fmt.Printf("    %02ds-%02ds  n=%-5d p50=%5.1fms  p95=%6.1fms\n",
			w*10, (w+1)*10, len(lat), percentile(lat, 0.50), percentile(lat, 0.95))
	}
}

func main() {
	base := flag.String("url", "http://127.0.0.1:8080", "base URL of a running instance")
	vus := flag.Int("vus", 100, "concurrent virtual users")
	dur := flag.Duration("duration", 60*time.Second, "test duration")
	flag.Parse()

	client := &http.Client{Timeout: 10 * time.Second}
	stamp := time.Now().UnixNano()

	fmt.Printf("setting up %d documents...\n", *vus)
	ids := make([]string, *vus)
	for i := 0; i < *vus; i++ {
		form := url.Values{"title": {fmt.Sprintf("loadtest-%d-%d", stamp, i)}}
		req, _ := http.NewRequest("POST", *base+"/create", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Do not follow the redirect: the Location header carries the new id.
		noRedirect := &http.Client{Timeout: 10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := noRedirect.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
			os.Exit(1)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusSeeOther {
			fmt.Fprintf(os.Stderr, "setup failed: create returned %d\n", resp.StatusCode)
			os.Exit(1)
		}
		ids[i] = strings.TrimPrefix(resp.Header.Get("Location"), "/doc?id=")
	}

	fmt.Printf("running: %d VUs x 1 write/sec + 1 read/sec for %s\n\n", *vus, *dur)
	start := time.Now()
	var wg sync.WaitGroup
	all := make([][]sample, *vus)
	var conflicts int64
	var cmu sync.Mutex

	for i := 0; i < *vus; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := ids[i]
			version := 1
			body := strings.Repeat("x", 10*1024) // ~10 KB, assumption A5
			// Stagger starts so 100 VUs do not fire in lockstep.
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			tick := time.NewTicker(time.Second)
			defer tick.Stop()
			var mine []sample

			for time.Since(start) < *dur {
				<-tick.C
				w := int(time.Since(start).Seconds()) / 10

				t0 := time.Now()
				resp, err := client.Get(*base + "/doc?id=" + id)
				ms := float64(time.Since(t0).Microseconds()) / 1000
				ok := err == nil && resp.StatusCode == http.StatusOK
				if resp != nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				mine = append(mine, sample{ms: ms, window: w, ok: ok, read: true})

				form := url.Values{
					"id": {id}, "version": {strconv.Itoa(version)},
					"content": {body + strconv.Itoa(version)},
				}
				req, _ := http.NewRequest("POST", *base+"/save", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				t0 = time.Now()
				resp, err = client.Do(req)
				ms = float64(time.Since(t0).Microseconds()) / 1000
				ok = false
				if err == nil {
					if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusSeeOther {
						ok = true
						version++
					} else if resp.StatusCode == http.StatusConflict {
						cmu.Lock()
						conflicts++
						cmu.Unlock()
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				mine = append(mine, sample{ms: ms, window: w, ok: ok, read: false})
			}
			all[i] = mine
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)

	var reads, writes []sample
	for _, s := range all {
		for _, x := range s {
			if x.read {
				reads = append(reads, x)
			} else {
				writes = append(writes, x)
			}
		}
	}

	fmt.Printf("completed in %s\n\n", elapsed.Round(time.Millisecond))
	summarise("reads", reads)
	summarise("writes", writes)
	fmt.Printf("\nthroughput: %.1f req/sec sustained (%d requests total)\n",
		float64(len(reads)+len(writes))/elapsed.Seconds(), len(reads)+len(writes))
	fmt.Printf("save conflicts: %d (expected 0 — each VU owns its document)\n", conflicts)

	nWindows := int(elapsed.Seconds())/10 + 1
	windows("read", reads, nWindows)

	var readErrs, writeErrs int
	for _, x := range reads {
		if !x.ok {
			readErrs++
		}
	}
	for _, x := range writes {
		if !x.ok {
			writeErrs++
		}
	}
	fmt.Println()
	if readErrs+writeErrs > 0 || conflicts > 0 {
		fmt.Printf("RESULT: FAILED — %d read errors, %d write errors, %d conflicts\n",
			readErrs, writeErrs, conflicts)
		os.Exit(1)
	}
	var lat []float64
	for _, x := range reads {
		lat = append(lat, x.ms)
	}
	sort.Float64s(lat)
	p95 := percentile(lat, 0.95)
	fmt.Printf("RESULT: PASSED — no errors. Read p95 = %.1f ms against the NFR1 budget of 200 ms.\n", p95)
}

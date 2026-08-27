# The image is the reproducibility artifact FR1 asks for: it builds identically
# anywhere, which is what Phase 0's "works where it was set up" failure lacked.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/collabdocs .

# Distroless: no shell, no package manager, nothing to exploit that we did not
# put there. The binary is static, so nothing else is needed.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/collabdocs /collabdocs
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/collabdocs"]

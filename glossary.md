# Glossary

Acronyms and jargon used across the course worksheets and decisions, with what they mean *here*.

## Course vocabulary (the worksheet's own language)

| Term | Expansion | What it means here |
|---|---|---|
| **SDLC** | Software Development Life Cycle | The phase list from worksheet 1: white paper → domain analysis → requirements → solution analysis → design → implementation → integration → testing → release → retirement |
| **FR** | Functional Requirement | What the system *does* — the verbs. "A user can create a document." Cited as FR1, FR2… |
| **NFR** | Non-Functional Requirement | How *well* it does it — the adverbs. Latency, recovery time, deploy time. These are what force architecture to exist |
| **QA** | **Quality Attribute** — *not* Quality Assurance | In "QA ranking: Performance > Resilience > Scalability > Consistency > Maintainability", QA = the -ilities. A tie-break order for when NFRs conflict. Easy trap |
| **MVP** | Minimum Viable Product | Smallest thing that proves the idea. Here defined numerically: ~1% of the year-1 goal = ~1,000 concurrent users |
| **ADR** | Architecture Decision Record | The industry name for what `decisions.md` is: one entry per decision, with context, choice, and consequences |
| **SDD** | Software Design Document | The elaborated design spec (Andrew's banking app used one) |

## Architecture & quality

| Term | Expansion | What it means here |
|---|---|---|
| **SPOF** | Single Point Of Failure | Any component whose death kills the whole system. One VM = one SPOF, no matter how many containers run on it |
| **HA** | High Availability | Designing so there's no downtime. Stronger promise than NFR3, which only asks for *recovery* < 60s |
| **AZ** | Availability Zone | A cloud provider's isolated datacenter. "Multi-AZ" = survives one datacenter dying |
| **p50 / p95** | 50th / 95th percentile | p95 < 200 ms means 95 of 100 requests finish under 200 ms. Percentiles are used instead of averages because averages hide the slow tail |
| **TTL** | Time To Live | How long a cached value stays valid before it's thrown away |
| **C10K** | "10,000 concurrent connections" | The classic problem of handling many simultaneous connections on one machine. Go solves it by default (goroutines + epoll), which is why compute isn't our bottleneck |
| **ACID** | Atomicity, Consistency, Isolation, Durability | The guarantees a relational database makes about transactions. Relevant to Decision 2 |

## Infrastructure & cloud

| Term | Expansion | What it means here |
|---|---|---|
| **VM** | Virtual Machine | A rented fake computer — virtualized hardware, you get root. qemu, industrialized |
| **AWS / GCP** | Amazon Web Services / Google Cloud Platform | The two cloud providers the worksheet names products from |
| **EC2** | Elastic Compute Cloud | AWS's VMs |
| **GCE** | Google Compute Engine | Google's VMs |
| **LB** | Load Balancer | A process owning the front port, forwarding each request to one of N instances |
| **ALB** | Application Load Balancer | AWS's managed HTTP load balancer |
| **ECS** | Elastic Container Service | AWS's container orchestrator |
| **K8s** | Kubernetes (K + 8 letters + s) | The industrial container orchestrator. Overkill at MVP scale |
| **FaaS** | Function as a Service | Serverless — you ship a handler, not a process. AWS Lambda, Google Cloud Functions |
| **IaC** | Infrastructure as Code | Declaring servers/networks in version-controlled files instead of clicking a console. Terraform, Pulumi, CloudFormation |
| **CDK** | Cloud Development Kit | AWS's IaC flavor where you write infrastructure in a real programming language |
| **CDN** | Content Delivery Network | Geographically distributed caches serving static content close to users |
| **PID** | Process ID | Containers get their own PID namespace — a private view of the process table |

## Software & dev practice

| Term | Expansion | What it means here |
|---|---|---|
| **XP** | Extreme Programming | Methodology defined by engineering practices: TDD, pairing, CI |
| **TDD** | Test-Driven Development | Write the failing test first |
| **CI** | Continuous Integration | Every commit automatically built and tested |
| **WIP** | Work In Progress | Kanban limits how many items are in flight at once |
| **CRUD** | Create, Read, Update, Delete | The four basic data operations |
| **DTO** | Data Transfer Object | A plain data shape for crossing a boundary, separate from the domain model. Phase 5 |
| **ORM** | Object-Relational Mapping | Library mapping DB rows to objects. Phase 3 territory — banned now (Constraint 1) |
| **API** | Application Programming Interface | The published surface one piece of software offers another |
| **REST** | Representational State Transfer | HTTP API style built around resources and verbs |
| **IPC** | Inter-Process Communication | How separate processes talk — signals, pipes, sockets |
| **IP** | Intellectual Property | Worksheet 1: concentrates in the Generic Infrastructure layer |
| **OTS** | Off-The-Shelf | The bought-components layer |

## Web & wire formats

| Term | Expansion | What it means here |
|---|---|---|
| **HTTP** | HyperText Transfer Protocol | Request/response protocol the app speaks |
| **HTML** | HyperText Markup Language | What the server renders and the browser displays |
| **URL** | Uniform Resource Locator | The address. Spaces get escaped as `%20` — visible in our `/doc?name=sprint%20notes` |
| **JSON** | JavaScript Object Notation | Text data format. Considered and rejected for metadata in D2 |
| **CRLF / LF** | Carriage Return + Line Feed (`\r\n`) / Line Feed (`\n`) | The line-ending bug we hit: browsers submit textarea content as CRLF; D2 says files hold LF |
| **ASCII** | American Standard Code for Information Interchange | The basic character encoding |

## Business / domain

| Term | Expansion | What it means here |
|---|---|---|
| **SaaS** | Software as a Service | Software rented per-period rather than sold as a copy |
| **CRM** | Customer Relationship Management | Worksheet 1's layering exercise |
| **EHS** | Environment, Health & Safety | Andrew's safety-compliance SaaS, used as the worked example in worksheet 1 |

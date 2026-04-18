# go-grpc-microservices-benchmark

A production-style microservices benchmark built with Go, gRPC, Redis, and PostgreSQL. Three services talk to each other over gRPC, sit behind a custom API Gateway with round-robin load balancing, and use Redis for cache-aside caching.

Built to explore how real-world microservice patterns hold up under load — not just as a toy example, but with actual concerns like cache invalidation, distributed transactions, JWT auth, and graceful shutdown.

---

## What's Inside

Three services, one gateway, one cache, one database.

```
Client
  └── API Gateway :8000          (routing, JWT auth, round-robin LB)
        ├── user-service :8080   (users, auth, JWT issuance)
        ├── order-service :8082  (orders, talks to user + product via gRPC)
        └── product-service :8081 (products, stock management)
              │
              ├── PostgreSQL     (persistent storage)
              └── Redis          (cache-aside, LRU eviction, TTL-based expiry)
```

**Service-to-service communication** happens over gRPC — order-service calls user-service to validate a user exists, and product-service to check and deduct stock. All of this bypasses the gateway entirely.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| Service communication | gRPC + Protocol Buffers |
| HTTP routing | chi v5 |
| Database | PostgreSQL 16 |
| Caching | Redis 7 (cache-aside, allkeys-lru) |
| Auth | JWT (HS256) |
| Containerization | Docker + Docker Compose |
| Logging | logrus |

---

## Project Structure

```
go-grpc-microservices-benchmark/
├── gateway/
│   ├── cmd/main.go             # gateway entry point
│   ├── router.go               # reverse proxy 
│   └── loadbalancer/           # Different load balancing strategies 
|                               #LeastConnections,Round-Robin,Consistent Hash
│
├── services/
│   ├── user/                   # handlers, repo, gRPC server
│   ├── order/                  # handlers, repo, gRPC clients
│   └── product/                # handlers, repo, gRPC server
│
├── proto/                      #Protocol buffers
│   ├── user/user.proto
│   ├── order/order.proto
│   └── product/product.proto
│
├── pb/                         # auto-generated — do not edit
│   ├── user/
│   ├── order/
│   └── product/
│
├── internal/
│   ├── config/                 # env-based config loading
│   ├── db/                     # postgres setup + migrations
│   ├── middleware/             # JWT verification middleware
│   ├── cache/                  # Redis client setup
│   └── jwt/                    # JWT helpers
|   └── etcd/                   # Service Registry and watcher 
│ 
├── simulator/                  # simulates skewed , uniform and burst traffic
|
|
├── analyzer/                   # Generates reports based on the traffic. 
|                               # Helps study p95 and p99 latency for
|                               # different load balancing algorithms    
|
|
└── .env.example
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker 
- `protoc` compiler (for regenerating proto files)

```bash
# Install protoc
sudo apt install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Add to PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Clone and Configure

```bash
git clone https://github.com/genin6382/go-grpc-microservices-benchmark.git
cd go-grpc-microservices-benchmark

cp .env.example .env
# Edit .env 
```
Services come up in this order: PostgreSQL → Redis → all three services → API Gateway.

### Run Locally 

```bash
# Create docker containers ( only once)
docker run -d --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  gcr.io/etcd-development/etcd:v3.5.15 \
  /usr/local/bin/etcd \
  --name node1 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380


docker run -d --name redis-cache -p 6379:6379 redis:7-alpine redis-server --appendonly yes --maxmemory 512mb --maxmemory-policy allkeys-lru --loglevel notice

docker run --name postgres -e POSTGRES_PASSWORD=<password> -p 5432:5432 -v pgdata:/var/lib/postgresql -d postgres

# Start only infrastructure
docker start redis-cache etcd postgres

# Run each service in separate terminals
go run ./user-service/cmd/main.go
go run ./order-service/cmd/main.go
go run ./product-service/cmd/main.go
go run ./gateway/cmd/main.go
```
---

## API Reference

All requests go through the API Gateway on `:8000`. Protected routes require a `Authorization: Bearer <token>` header.

### Auth (no token required)
```
POST /users/login       — get JWT token
POST /users/register    — create account
```

### Users (protected)
```
GET    /users/          — list all users
GET    /users/{id}      — get user by ID
DELETE /users/{id}      — delete user
```

### Products (protected)
```
GET    /products/       — list all products
GET    /products/{id}   — get product by ID
POST   /products/       — create product
PATCH  /products/{id}   — update product details
DELETE /products/{id}   — delete product
```

### Orders (protected)
```
GET    /orders/                    — list all orders
GET    /orders/{id}                — get order by ID
GET    /orders/user/{user_id}      — get orders for a user
POST   /orders/                    — place an order
PATCH  /orders/{id}                — update order status
DELETE /orders/{id}                — delete order
```

---

## Caching Strategy

Redis is used as a **cache-aside** layer across product and order services.

| Key Pattern | TTL | Invalidated When |
|---|---|---|
| `product:{id}` | 5 min | Product updated, stock deducted |
| `products:list` | 1 min | Any product write |
| `order:{id}` | 30 min | Order status updated |
| `orders:user:{id}` | 2 min | Any order write for that user |

Cache invalidation is always synchronous and happens **after** the DB write succeeds. Cache failures are logged but never return errors to the client — the app falls back to the database gracefully.

---

## gRPC Communication

Order-service makes two internal gRPC calls when placing an order:

```
POST /orders/
  └── CheckUserExists  → user-service:50051
  └── DeductStock      → product-service:50052
  └── CreateOrder      → order-service DB
```

Proto definitions live in `proto/` and generated code in `pb/`. To regenerate after editing a `.proto` file:

```bash
 protoc --go_out=. --go-grpc_out=. proto/<service>/<service>.proto
```

---

## Load Balancing

The API Gateway uses **round-robin** load balancing across service instances using Go's `httputil.ReverseProxy`. Each service can have any number of instances — just add more URLs to the `*_SERVICE_URLS` env vars.

gRPC connections between services also use round-robin via gRPC's built-in `round_robin` load balancing policy across multiple addresses.

---

## Health Checks

```bash
# Gateway health
curl http://localhost:8000/health

# Individual services (bypass gateway)
curl http://localhost:8080/health   # user-service
curl http://localhost:8081/health   # product-service
curl http://localhost:8082/health   # order-service

# Redis
docker exec redis redis-cli ping    # should return PONG
```

---

## Design Decisions

**Why gRPC for internal calls?**
Binary serialization, strict typed contracts, and lower overhead compared to REST for service-to-service calls. The `.proto` file acts as a living API contract between services.

**Why cache-aside over write-through?**
Only requested data gets cached, cache failures degrade gracefully to the DB, and the cache model can differ from the DB schema. Write-through was overkill for this read-heavy workload.

**Why a custom gateway over Nginx/Kong?**
Full control over load balancing logic, easy to extend with custom middleware (rate limiting, tracing), and it fits naturally into the Go codebase without extra ops overhead.

**Why keep gRPC internal?**
gRPC uses HTTP/2 and binary frames — browsers can't consume it natively without a proxy. Keeping it internal means services communicate efficiently while external clients use clean REST over the gateway.

---

## Contributing

This is a benchmark/learning project but PRs are welcome. If you find a caching bug, a race condition, or a better pattern — open an issue.

---


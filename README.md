# Full Observability Microservices

Bu proje, Go ile yazılmış mikroservis mimarisi ve tam observability (gözlemlenebilirlik) özelliklerini içerir.

## Servisler

### 1. User Service (Port: 8080, 9090)
- Kullanıcı yönetimi (CRUD)
- gRPC ve HTTP desteği
- CQRS pattern
- JWT authentication

### 2. Product Service (Port: 8081)
- Ürün yönetimi (CRUD)
- **CQRS Pattern** (Command Query Responsibility Segregation)
- **JWT Authentication** (Admin endpoints protected)
- **Role-based Authorization** (Admin/Public endpoints)
- Stok yönetimi
- Kategori bazlı filtreleme
- İstatistik endpoint'i
- REST API

## Observability Stack

- **Prometheus** (Port: 9091) - Metrics toplama
- **Grafana** (Port: 3000) - Görselleştirme (admin/admin)
- **Jaeger** (Port: 16686) - Distributed tracing
- **PostgreSQL** (Port: 5432) - Veritabanı

```mermaid
graph TB
    subgraph "Docker Containers"
        US[User Service<br/>:8080, :9090]
        PS[Product Service<br/>:8081]
        PG[(PostgreSQL<br/>:5432)]
        PR[Prometheus<br/>:9091]
        GR[Grafana<br/>:3000]
        JG[Jaeger<br/>:16686]
    end
    
    US -->|userdb| PG
    PS -->|productdb| PG
    PR -->|Scrape /metrics<br/>every 10s| US
    PR -->|Scrape /metrics<br/>every 10s| PS
    GR -->|PromQL Query| PR
    US -->|Traces| JG
    PS -->|Traces| JG
    
    Client([Client]) -->|HTTP/gRPC| US
    Client -->|HTTP| PS
    Admin([Admin]) -->|View Dashboard| GR
    Admin -->|View Traces| JG
```

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler
    participant M as Metrics Middleware
    participant S as Service
    participant R as Repository
    participant DB as PostgreSQL
    participant P as Prometheus

    C->>H: POST /users
    H->>M: metricsMiddleware()
    M->>M: start = time.Now()
    M->>S: CreateUser()
    S->>S: Validate
    S->>R: Create()
    R->>DB: INSERT INTO users
    DB-->>R: User ID
    R-->>S: User
    S-->>M: User
    M->>M: duration = time.Since(start)
    M->>M: requestLatency.Observe(duration)
    M->>M: requestCounter.Inc()
    M-->>H: Response
    H-->>C: 201 Created
    
    P->>H: GET /metrics
    H-->>P: Prometheus Format Data
```

## Hızlı Başlangıç

### Tüm servisleri çalıştır
```bash
make docker-up
```

### Servisleri test et
```bash
# Product API test
./scripts/test-product-api.sh
```

### Servisleri durdur
```bash
make docker-down
```

### Yerel geliştirme
```bash
# User service
make run-user

# Product service
make run-product
```

## Endpoints

### User Service
- HTTP: http://localhost:8080
- gRPC: localhost:9090
- Swagger: http://localhost:8080/swagger/

### Product Service
- HTTP: http://localhost:8081
- Health: http://localhost:8081/health

### Monitoring
- Prometheus: http://localhost:9091
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686

```mermaid
graph LR
    subgraph "Project Structure"
        direction TB
        A[cmd/]
        B[internal/]
        C[pkg/]
        D[prometheus/]
        E[grafana/]
        F[dockerfiles/]
        
        A --> A1[user/main.go]
        A --> A2[product/main.go]
        
        B --> B1[user/<br/>CQRS Pattern]
        B --> B2[product/<br/>CQRS Pattern]
        
        C --> C1[database/]
        C --> C2[logger/]
        C --> C3[tracing/]
        C --> C4[auth/]
    end
```

```mermaid
graph LR
    subgraph "CQRS Pattern - Product Service"
        direction TB
        H[HTTP Handler]
        
        subgraph Commands
            C1[CreateProductCommand]
            C2[UpdateProductCommand]
            C3[DeleteProductCommand]
            C4[UpdateStockCommand]
        end
        
        subgraph Queries
            Q1[GetProductQuery]
            Q2[ListProductsQuery]
            Q3[GetStatsQuery]
        end
        
        R[(Repository)]
        
        H -->|Write Operations| Commands
        H -->|Read Operations| Queries
        Commands --> R
        Queries --> R
    end
```

```mermaid
graph TD
    subgraph "Prometheus Metrics"
        M1[user_service_requests_total<br/>CounterVec]
        M2[user_service_request_duration_seconds<br/>HistogramVec]
        M3[user_service_active_users<br/>Gauge]
        
        M1 --> L1[Labels: method, endpoint, status]
        M2 --> L2[Labels: method, endpoint]
        M3 --> L3[No Labels]
    end
```

```mermaid
graph LR
    subgraph "API Endpoints"
        E1[POST /users<br/>Create User]
        E2[GET /users<br/>List All Users]
        E3[GET /users/:id<br/>Get User]
        E4[PUT /users/:id<br/>Update User]
        E5[DELETE /users/:id<br/>Delete User]
        E6[GET /health<br/>Health Check]
        E7[GET /metrics<br/>Prometheus]
    end
```

```mermaid
flowchart TD
    Start([docker-compose up]) --> DB[PostgreSQL Start]
    DB --> HC{Health Check}
    HC -->|Ready| US[User Service Start]
    HC -->|Wait| HC
    US --> Schema[InitSchema<br/>CREATE TABLE]
    Schema --> Metrics[Register Prometheus Metrics]
    Metrics --> Router[Setup HTTP Router]
    Router --> Listen[Listen :8080]
    
    P[Prometheus Start] --> Config[Load prometheus.yml]
    Config --> Scrape[Start Scraping<br/>every 10s]
    
    G[Grafana Start] --> DS[Auto-provision<br/>Prometheus Datasource]
    DS --> Ready([System Ready])
```

```mermaid
graph TB
    subgraph "Layer Architecture"
        direction TB
        H[Handler Layer<br/>HTTP + Metrics]
        S[Service Layer<br/>Business Logic]
        R[Repository Layer<br/>Data Access]
        D[(Database)]
        
        H -->|Calls| S
        S -->|Calls| R
        R -->|SQL| D
    end
```

```mermaid
timeline
    title Observability Data Flow
    Request Comes : Client sends HTTP request
    Metrics Start : Middleware records start time
    Processing : Handler → Service → Repository → DB
    Metrics End : Middleware calculates duration
    Metrics Update : Counter++, Histogram.Observe(), Gauge.Set()
    Prometheus Scrape : Prometheus reads /metrics endpoint
    Grafana Query : Admin queries metrics via PromQL
    Visualization : Dashboard shows real-time graphs
```


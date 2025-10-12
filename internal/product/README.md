# Product Service

Basit bir ürün yönetim servisi. REST API ile ürün CRUD işlemlerini destekler.

## Özellikler

- ✅ Ürün oluşturma, okuma, güncelleme ve silme (CRUD)
- ✅ Kategori bazlı filtreleme
- ✅ Stok yönetimi
- ✅ Health check endpoint
- ✅ Prometheus metrics
- ✅ Distributed tracing (Jaeger)
- ✅ Structured logging

## API Endpoints

### Ürün Oluşturma
```bash
POST /api/products
Content-Type: application/json

{
  "name": "iPhone 15",
  "description": "Latest iPhone model",
  "price": 999.99,
  "stock": 100,
  "category": "Electronics",
  "sku": "IPH15-001",
  "is_active": true
}
```

### Ürün Listesi
```bash
GET /api/products?limit=10&offset=0&category=Electronics
```

### Ürün Detayı
```bash
GET /api/products/{id}
```

### Ürün Güncelleme
```bash
PUT /api/products/{id}
Content-Type: application/json

{
  "name": "iPhone 15 Pro",
  "price": 1199.99
}
```

### Ürün Silme
```bash
DELETE /api/products/{id}
```

### Stok Güncelleme
```bash
PATCH /api/products/{id}/stock
Content-Type: application/json

{
  "stock": 50
}
```

### Health Check
```bash
GET /health
```

### Metrics
```bash
GET /metrics
```

## Çalıştırma

### Yerel olarak
```bash
# Önce PostgreSQL'in çalıştığından emin olun
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=productdb
export HTTP_PORT=8081

make run-product
```

### Docker ile
```bash
# Tüm servisleri çalıştır
make docker-up

# Sadece product servisi
docker-compose up product-service
```

## Veritabanı

Service otomatik olarak `products` tablosunu oluşturur (auto-migration).

### Product Schema
- `id` - Primary Key
- `name` - Ürün adı
- `description` - Ürün açıklaması
- `price` - Fiyat
- `stock` - Stok miktarı
- `category` - Kategori
- `sku` - Stok Kodu (Unique)
- `is_active` - Aktif durumu
- `created_at` - Oluşturma zamanı
- `updated_at` - Güncelleme zamanı
- `deleted_at` - Soft delete zamanı

## Monitoring

- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3000
- **Jaeger**: http://localhost:16686


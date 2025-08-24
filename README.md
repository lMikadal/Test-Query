# Benchmark Testing Project

โปรเจกต์สำหรับทดสอบประสิทธิภาพระหว่าง Redis และ Split Database ด้วยข้อมูล 300,000 records

## Tech Stack

- **Backend**: Golang Echo Framework
- **Database**: PostgreSQL
- **Cache**: Redis
- **Migration**: Goose
- **Container**: Docker & Docker Compose

## Project Structure

```
├── docker-compose.yml          # Docker Compose configuration
├── Dockerfile                  # Golang application container
├── migrations/                 # Database migration files
│   └── 001_create_benchmark_table.sql
├── scripts/                    # Database initialization scripts
│   └── init-db.sql
├── pgadmin/                    # pgAdmin configuration
│   └── servers.json           # Pre-configured database servers
├── logs/                       # Application logs
├── env.example                 # Environment variables example
└── .gitignore                  # Git ignore file
```

## Architecture

### Database Setup
1. **postgres_main**: Main PostgreSQL database (port 5432)
2. **redis**: Redis cache (port 6379)
3. **postgres_read**: Read-only database for split DB testing (port 5433)
4. **postgres_write**: Write-only database for split DB testing (port 5434)

### Services
- **app**: Golang Echo application (port 8080)
- **goose**: Migration tool
- **pgadmin**: Database management UI (port 5050)

## Quick Start

1. **Clone และ Setup**
   ```bash
   git clone <repository>
   cd Test-Query
   cp env.example .env
   ```

2. **Start All Services**
   ```bash
   docker-compose up -d
   ```

3. **Check Services Status**
   ```bash
   docker-compose ps
   ```

4. **View Logs**
   ```bash
   docker-compose logs -f app
   ```

## Database Access

### PostgreSQL Databases
- **Main DB**: `localhost:5432`
- **Read DB**: `localhost:5433`
- **Write DB**: `localhost:5434`
- **Username**: `postgres`
- **Password**: `postgres123`

### Redis
- **Host**: `localhost:6379`

### pgAdmin
- **URL**: `http://localhost:5050`
- **Email**: `admin@benchmark.com`
- **Password**: `admin123`
- **Pre-configured servers**: Main DB, Read DB, Write DB
- **Auto-connect**: Servers จะถูก configure อัตโนมัติ

## Testing Scenarios

### 1. Redis Caching Test
- ใช้ Redis เป็น cache layer
- ทดสอบ read performance จาก cache
- เปรียบเทียบกับการ query ตรงจาก database

### 2. Split Database Test
- แยก read operations ไป `postgres_read`
- แยก write operations ไป `postgres_write`
- ทดสอบประสิทธิภาพของ load balancing

## Benchmark Data

- **Records**: 300,000 rows
- **Table**: `benchmark_data`
- **Fields**: id, name, email, age, city, country, timestamps, random_number, description
- **Indexes**: Optimized for common query patterns

## Commands

### Start Services
```bash
# Start all services
docker-compose up -d

# Start specific service
docker-compose up -d postgres_main redis
```

### Migration
```bash
# Run migrations
docker-compose run --rm goose

# Create new migration
docker-compose run --rm goose goose create migration_name sql
```

### Monitoring
```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f app
docker-compose logs -f postgres_main
```

### Database Management with pgAdmin
```bash
# เข้าใช้งาน pgAdmin
# เปิด browser ไปที่ http://localhost:5050
# Login ด้วย admin@benchmark.com / admin123
# Servers จะถูก configure อัตโนมัติ (Main DB, Read DB, Write DB)

# หาก servers ไม่แสดง ให้ restart pgAdmin
docker-compose restart pgadmin
```

### Cleanup
```bash
# Stop all services
docker-compose down

# Remove volumes (delete data)
docker-compose down -v
```

## Environment Variables

สำคัญ: คัดลอก `env.example` เป็น `.env` และปรับค่าตามความต้องการ

```bash
cp env.example .env
```

## Development

1. **Build Application**
   ```bash
   docker-compose build app
   ```

2. **Restart Application**
   ```bash
   docker-compose restart app
   ```

3. **Enter Container**
   ```bash
   docker-compose exec app sh
   ```

## Performance Testing

โปรเจกต์นี้ออกแบบมาเพื่อทดสอบ:

1. **Latency**: เวลาที่ใช้ในการ response
2. **Throughput**: จำนวน requests ที่สามารถจัดการได้ต่อวินาที
3. **Resource Usage**: การใช้ CPU, Memory, และ Network
4. **Concurrency**: ประสิทธิภาพภายใต้ concurrent requests

## Next Steps

1. สร้าง Golang Echo application
2. Implement benchmark endpoints
3. สร้าง test data 300,000 records
4. ทดสอบประสิทธิภาพและเปรียบเทียบผลลัพธ์
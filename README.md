# Triply Server

A Go-based backend API server for the Triply travel planning application, built with Fiber (web framework) and GORM (ORM).

## Features

- ✅ **Trip Management** - CRUD operations for trips, destinations, daily plans, and activities
- ✅ **Public Trips** - Share trips publicly with advanced filtering and search
- ✅ **Google OAuth** - Secure authentication via Google
- ✅ **JWT Authentication** - Token-based auth with httpOnly cookies
- ✅ **Shadow Users** - Anonymous trip creation before login
- ✅ **Activity Ordering** - Persist drag-and-drop activity reordering
- ✅ **Trip Likes** - Like/unlike public trips
- ✅ **Trip Import** - Import parts of public trips
- ✅ **PostgreSQL** - Production-ready database
- ✅ **CORS** - Configured for Next.js frontend
- ✅ **Layered Architecture** - Clean separation of concerns

---

## Quick Start

### Prerequisites

- **Go 1.23+**
- **PostgreSQL 14+** (or SQLite for development)

### Installation

1. **Create `.env` file** (copy from example):
```bash
cp .env.example .env
```

2. **Configure environment variables** (see below)

3. **Build and run:**
   ```bash
   go build -o bin/server ./cmd/server
   ./bin/server
   ```

Server runs on **http://localhost:8080**

---

## Environment Variables

Create a `.env` file in the project root with these variables:

### Required Configuration

```bash
# Server Configuration
PORT=8080
FRONTEND_ORIGIN=http://localhost:3001

# Database Configuration (PostgreSQL)
DATABASE_URL=host=localhost user=triply_user password=triply_local_dev dbname=triply_dev port=5432 sslmode=disable

# JWT Configuration
JWT_SECRET=dev-secret-change-me-in-production
```

### Google OAuth (Required for Login)

```bash
# Google OAuth Configuration
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-your-client-secret
OAUTH_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

**To get Google OAuth credentials:**

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project (or use existing)
3. Enable Google+ API
4. Create OAuth 2.0 Client ID credentials
5. Configure:
   - **Authorized JavaScript origins:** `http://localhost:3001`, `http://localhost:8080`
   - **Authorized redirect URIs:** `http://localhost:8080/auth/google/callback`
6. Copy Client ID and Client Secret to `.env`

---

## Running Locally

### Development Mode

```bash
# Build and run
go build -o bin/server ./cmd/server && ./bin/server
```

### With Hot Reload (Recommended)

Install Air:
```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:
```bash
air
```

The server will automatically restart when you save Go files.

---

## API Documentation

### Base URL
```
http://localhost:8080
```

### Authentication Endpoints

#### Google OAuth Login
```http
GET /auth/google
```
Redirects to Google's OAuth login page.

#### OAuth Callback
```http
GET /auth/google/callback?code=...
```
Handles Google OAuth callback. Sets JWT token in httpOnly cookie.

#### Get Current User
```http
GET /auth/me
Cookie: triply_token=<jwt-token>
```
Returns current authenticated user or null.

#### Logout
```http
POST /auth/logout
Cookie: triply_token=<jwt-token>
```
Clears authentication cookies.

#### Migrate Shadow Trips
```http
POST /auth/migrate-shadow-trips
Body: { "shadowUserId": "shadow-xxx" }
```
Migrates anonymous trips to authenticated user account.

### Trip Endpoints

#### List User Trips
```http
GET /api/users/:userId/trips
Header: X-Shadow-User-ID: shadow-xxx (for anonymous users)
```

#### Create Trip
```http
POST /api/users/:userId/trips
Body: { "trip": {...} }
```

#### Update Trip
```http
PUT /api/users/:userId/trips/:tripId
Body: { "trip": {...} }
```

#### Delete Trip
```http
DELETE /api/users/:userId/trips/:tripId
```

### Public Trip Endpoints

#### List Public Trips (with Filters)
```http
GET /api/public-trips?page=1&pageSize=50&sort=featured&filters={"months":[1,2],"durations":[...],"travelerTypes":["solo"],"cities":["Tokyo"]}
```

**Query Parameters:**
- `page` - Page number (default: 1)
- `pageSize` - Items per page (default: 50)
- `sort` - Sort key: `featured`, `likes`, `recent` (default: featured)
- `filters` - JSON string with filter criteria

**Filter Options:**
- `months` - Array of months (1-12)
- `durations` - Array of duration ranges: `[{minDays: 1, maxDays: 7}]`
- `travelerTypes` - Array: `["solo", "couple", "family", "friends"]`
- `cities` - Array of city names

#### Get Public Trip Detail
```http
GET /api/public-trips/:tripId
```

#### Toggle Trip Visibility
```http
POST /api/public-trips/:tripId/visibility
Body: { "visibility": "public" | "private" }
```

#### Like/Unlike Trip
```http
POST /api/public-trips/:tripId/like
```
Toggles like status for authenticated user.

### Activity Endpoints

#### Update Activity Order
```http
POST /api/activities/order
Body: { "tripId": "...", "dayId": "...", "activities": [...] }
```

### Health Check

```http
GET /api/health
Response: {"status": "healthy", "database": "up"}
```

---

## Project Structure

```
triply-server/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── models/                  # Database models
│   │   ├── user.go
│   │   ├── trip.go
│   │   ├── types.go
│   │   └── trip_like.go
│   ├── repository/              # Data access layer
│   │   ├── user_repository.go
│   │   ├── trip_repository.go
│   │   ├── public_trip_repository.go
│   │   └── trip_like_repository.go
│   ├── service/                 # Business logic
│   │   ├── auth_service.go
│   │   ├── trip_service.go
│   │   ├── public_trip_service.go
│   │   └── trip_like_service.go
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth_handler.go
│   │   ├── trip_handler.go
│   │   ├── public_trip_handler.go
│   │   └── trip_like_handler.go
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logger.go
│   │   └── error.go
│   ├── dto/                     # Data transfer objects
│   │   ├── trip_dto.go
│   │   └── public_trip_dto.go
│   └── utils/                   # Utility functions
├── bin/                         # Compiled binaries
├── .env                         # Environment variables (gitignored)
├── .env.example                 # Example configuration
├── Makefile                     # Build commands
└── README.md                    # This file
```

---

## Configuration Details

### Development `.env`

```bash
# Server
PORT=8080
FRONTEND_ORIGIN=http://localhost:3001

# Database (PostgreSQL for development)
DATABASE_URL=host=localhost user=triply_user password=triply_local_dev dbname=triply_dev port=5432 sslmode=disable

# JWT
JWT_SECRET=dev-secret-change-me

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-your-client-secret
OAUTH_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

### Production `.env`

```bash
# Server
PORT=8080
FRONTEND_ORIGIN=https://yourdomain.com

# Database (PostgreSQL)
DATABASE_URL=postgresql://user:password@host:5432/triply_prod?sslmode=require

# JWT (use strong secret!)
JWT_SECRET=<generate-secure-random-string>

# Google OAuth
GOOGLE_CLIENT_ID=your-prod-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-your-prod-client-secret
OAUTH_REDIRECT_URL=https://api.yourdomain.com/auth/google/callback

# Optional
GO_ENV=production
```

**Generate secure JWT secret:**
```bash
openssl rand -base64 32
```

---

## Database Setup

### PostgreSQL (Recommended)

1. **Install PostgreSQL**
   ```bash
   brew install postgresql@14
   brew services start postgresql@14
   ```

2. **Create database and user**
   ```bash
   psql postgres
   ```
   ```sql
   CREATE DATABASE triply_dev;
   CREATE USER triply_user WITH PASSWORD 'triply_local_dev';
   GRANT ALL PRIVILEGES ON DATABASE triply_dev TO triply_user;
   ```

3. **Configure in `.env`**
   ```bash
   DATABASE_URL=host=localhost user=triply_user password=triply_local_dev dbname=triply_dev port=5432 sslmode=disable
   ```

### Auto-Migration

The server automatically:
- ✅ Drops old tables on startup (development only)
- ✅ Creates new tables with correct schema
- ✅ Seeds demo data (3 public trips with full itineraries)

---

## Development Commands

```bash
# Build
make build              # Build binary to ./bin/server
go build -o bin/server ./cmd/server

# Run
make run                # Build and run
./bin/server            # Run compiled binary

# Development (with hot reload)
make dev                # Requires air installed
air                     # Uses .air.toml configuration

# Testing
make test               # Run all tests
make test-coverage      # Run tests with coverage report

# Code Quality
make fmt                # Format code (gofmt)
make lint               # Run golangci-lint
make vet                # Run go vet

# Cleanup
make clean              # Remove compiled binaries
```

---

## Architecture

### Layered Design

```
HTTP Request
     ↓
[Middleware] → Auth, CORS, Logger, Error Handler
     ↓
[Handlers] → Parse request, validate input
     ↓
[Services] → Business logic, orchestration
     ↓
[Repositories] → Database queries (GORM)
     ↓
[Models] → Database entities
     ↓
PostgreSQL Database
```

### Key Patterns

- **Dependency Injection** - Services and repositories injected into handlers
- **Interface-Based** - All services/repositories use interfaces
- **Error Handling** - Centralized error middleware
- **Request/Response DTOs** - Clean API contracts

---

## Security

### Authentication

- **JWT Tokens** - Stored in httpOnly cookies (protected from XSS)
- **CSRF Protection** - SameSite cookie policy
- **Shadow Users** - Anonymous users get temporary IDs
- **Migration** - Shadow trips automatically migrate on login

### CORS

- Configured for specific frontend origin
- Credentials allowed for cookie support
- Proper headers for preflight requests

### Database

- **Prepared Statements** - GORM protects against SQL injection
- **Parameterized Queries** - No string concatenation
- **Foreign Key Constraints** - Data integrity enforced

---

## Deployment

### Using Docker (Coming Soon)

```dockerfile
# Dockerfile example
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/server ./cmd/server

FROM alpine:latest
COPY --from=builder /app/bin/server /server
CMD ["/server"]
```

### Direct Deployment

1. **Build for production:**
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server
   ```

2. **Set production environment variables**

3. **Run with systemd** (example):
   ```ini
   [Unit]
   Description=Triply API Server
   After=network.target

   [Service]
   Type=simple
   User=triply
   WorkingDirectory=/opt/triply-server
   EnvironmentFile=/opt/triply-server/.env
   ExecStart=/opt/triply-server/bin/server
   Restart=always

   [Install]
   WantedBy=multi-user.target
   ```

4. **Enable and start:**
   ```bash
   sudo systemctl enable triply-server
   sudo systemctl start triply-server
   ```

---

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -ti:8080 | xargs kill -9
```

### Database Connection Issues

1. Check PostgreSQL is running:
   ```bash
   psql -U triply_user -d triply_dev
   ```

2. Verify `DATABASE_URL` in `.env`

3. Check logs for specific error messages

### CORS Errors

1. Verify `FRONTEND_ORIGIN` matches your frontend URL
   ```bash
   # Development
   FRONTEND_ORIGIN=http://localhost:3001
   
   # Production
   FRONTEND_ORIGIN=https://yourdomain.com
   ```

2. Restart server after changing `.env`

### Google OAuth Errors

1. **"OAuth not configured"**
   - Add `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` to `.env`
   - Restart server

2. **"Invalid client"**
   - Verify Client ID and Secret are correct
   - Check Google Cloud Console OAuth client is active

3. **"Redirect URI mismatch"**
   - Add `http://localhost:8080/auth/google/callback` to authorized redirect URIs in Google Console

### Authentication Not Working

1. Check cookies are being set (browser DevTools → Application → Cookies)
2. Verify JWT secret is set in `.env`
3. Check frontend is sending `credentials: 'include'` in fetch requests

---

## API Examples

### List Public Trips with Filters

```bash
curl 'http://localhost:8080/api/public-trips?page=1&pageSize=10&sort=featured&filters=%7B%22months%22%3A%5B3%2C4%5D%2C%22durations%22%3A%5B%7B%22minDays%22%3A7%2C%22maxDays%22%3A14%7D%5D%2C%22travelerTypes%22%3A%5B%22solo%22%5D%7D'
```

### Get Trip Detail

```bash
curl http://localhost:8080/api/public-trips/pt-tokyo-week-discovery
```

### Health Check

```bash
curl http://localhost:8080/api/health
# Response: {"database":"up","status":"healthy"}
```

---

## Database Schema

### Main Tables

- **users** - User accounts (Google OAuth)
- **trips** - User trips with visibility settings
- **destinations** - Trip destinations
- **day_plans** - Daily activity plans
- **activities** - Individual activities with geolocation
- **trip_likes** - User likes on public trips

### Relationships

```
User ──< Trip ──< Destination ──< DayPlan ──< Activity
User ──< TripLike >── Trip
```

---

## Development Workflow

### Making Changes

1. **Edit code** in `internal/`
2. **Hot reload** automatically restarts (if using air)
3. **Test** with `make test`
4. **Format** with `make fmt`
5. **Lint** with `make lint`

### Adding New Endpoint

1. Create DTO in `internal/dto/`
2. Add repository method in `internal/repository/`
3. Add service method in `internal/service/`
4. Add handler in `internal/handlers/`
5. Register route in `cmd/server/main.go`

### Database Migrations

GORM AutoMigrate runs on startup:
- Detects schema changes
- Creates/alters tables automatically
- **Development:** Drops and recreates tables
- **Production:** Only adds new columns (safer)

---

## Frontend Integration

### CORS Configuration

The server is configured to work with the Next.js frontend:

```go
// Configured via FRONTEND_ORIGIN environment variable
middleware.NewCORS(cfg.Server.FrontendOrigin)
```

**Development:** `http://localhost:3001`  
**Production:** `https://yourdomain.com`

### Authentication Flow

1. Frontend redirects to `/auth/google`
2. Backend redirects to Google OAuth
3. User logs in with Google
4. Google redirects to `/auth/google/callback`
5. Backend creates/updates user, generates JWT
6. Backend sets httpOnly cookie with JWT
7. Backend redirects to frontend
8. Frontend reads auth state from `/auth/me`

### Shadow User System

**Anonymous Users:**
- Frontend generates shadow user ID: `shadow-{timestamp}-{random}`
- Stored in localStorage
- Sent in `X-Shadow-User-ID` header
- Allows trip creation before login

**After Login:**
- Shadow trips automatically migrate to user account
- Shadow ID cleared from localStorage
- User owns all previous anonymous trips

---

## Production Deployment

### Checklist

- [ ] Set strong `JWT_SECRET` (32+ random bytes)
- [ ] Use PostgreSQL (not SQLite)
- [ ] Configure production `FRONTEND_ORIGIN`
- [ ] Set up Google OAuth for production domain
- [ ] Enable HTTPS/TLS
- [ ] Set up reverse proxy (nginx/caddy)
- [ ] Configure systemd service
- [ ] Set up logging and monitoring
- [ ] Database backups
- [ ] Rate limiting (consider adding middleware)

### Environment Variables for Production

```bash
PORT=8080
FRONTEND_ORIGIN=https://yourdomain.com
DATABASE_URL=postgresql://user:pass@host:5432/triply_prod?sslmode=require
JWT_SECRET=<secure-random-string>
GOOGLE_CLIENT_ID=prod-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-prod-secret
OAUTH_REDIRECT_URL=https://api.yourdomain.com/auth/google/callback
GO_ENV=production
```

### Nginx Reverse Proxy Example

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Testing

### Run Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Verbose
go test -v ./...

# Specific package
go test ./internal/service/...
```

### Test Coverage

```bash
make test-coverage
# Opens coverage report in browser
```

---

## Monitoring

### Logs

The server logs all HTTP requests with:
- Method and path
- Client IP
- Status code
- Response time

### Health Endpoint

Monitor server health:
```bash
curl http://localhost:8080/api/health
```

Returns:
- `status`: `healthy` or `unhealthy`
- `database`: `up` or `down`

---

## Contributing

1. Follow the layered architecture
2. Write tests for new features
3. Use dependency injection
4. Keep handlers thin (business logic in services)
5. Use interfaces for testability
6. Format code: `make fmt`
7. Run linter: `make lint`
8. Update documentation

---

## License

Internal use only.

---

**Last Updated:** October 2025  
**Version:** 2.0.0

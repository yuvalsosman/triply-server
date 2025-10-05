# Triply Server

A Go-based backend server for the Triply travel planning application, built with Fiber and GORM.

## Features

- ✅ **Trip Management** - CRUD operations for trips, destinations, daily plans, and activities
- ✅ **Public Trips** - Share trips publicly with filtering and search
- ✅ **Google OAuth** - Secure authentication via Google
- ✅ **JWT Authentication** - Token-based auth with refresh support
- ✅ **Activity Ordering** - Persist activity reordering
- ✅ **Import System** - Import parts of public trips
- ✅ **SQLite & PostgreSQL** - Support for both databases
- ✅ **Layered Architecture** - Clean separation of concerns

## Project Structure

```
triply-server/
├── cmd/
│   └── server/         # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── models/         # Database models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic layer
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── dto/            # Data transfer objects
│   └── utils/          # Utility functions
├── bin/                # Compiled binaries
├── .env.example        # Example environment variables
├── Makefile            # Build and dev commands
└── README.md           # This file
```

## Quick Start

### Prerequisites

- Go 1.23 or higher
- SQLite (included) or PostgreSQL (optional)

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

3. Install dependencies:

```bash
make install-deps
```

4. Run the server:

```bash
make run
```

The server will start on `http://localhost:8080` by default.

### Development with Hot Reload

Install Air for hot reload:

```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:

```bash
make dev
```

## Configuration

Configure the server via environment variables in `.env`:

- `PORT` - Server port (default: 8080)
- `FRONTEND_ORIGIN` - Frontend URL for CORS
- `USE_SQLITE` - Use SQLite (1) or PostgreSQL (0)
- `SQLITE_PATH` - Path to SQLite database file
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret for JWT token signing
- `GOOGLE_CLIENT_ID` - Google OAuth client ID (optional)
- `GOOGLE_CLIENT_SECRET` - Google OAuth secret (optional)

## API Endpoints

### Authentication
- `GET /auth/google` - Initiate Google OAuth
- `GET /auth/google/callback` - OAuth callback
- `POST /auth/dev-login` - Dev login (development only)
- `GET /auth/me` - Get current user
- `POST /auth/logout` - Logout

### Trips
- `GET /api/users/:userId/trips` - List user trips
- `POST /api/users/:userId/trips` - Create trip
- `PUT /api/users/:userId/trips/:tripId` - Update trip
- `DELETE /api/users/:userId/trips/:tripId` - Delete trip

### Public Trips
- `GET /api/public-trips` - List public trips with filters
- `GET /api/public-trips/:tripId` - Get public trip details
- `POST /api/public-trips/:tripId/visibility` - Toggle visibility

### Activities
- `POST /api/activities/order` - Update activity order

### Import
- `POST /api/import-trip` - Import trip parts

### Health
- `GET /api/health` - Health check

## Development Commands

```bash
make build          # Build the server
make run            # Build and run
make dev            # Run with hot reload
make test           # Run tests
make test-coverage  # Run tests with coverage
make fmt            # Format code
make lint           # Run linter
make clean          # Clean build artifacts
```

## Architecture

The application follows a layered architecture:

1. **Handlers** - HTTP request/response handling
2. **Services** - Business logic
3. **Repositories** - Data access
4. **Models** - Database entities

### Middleware

- **Auth** - JWT token validation
- **CORS** - Cross-origin resource sharing
- **Logger** - Request logging
- **Error** - Centralized error handling

## Database

The server supports both SQLite (development) and PostgreSQL (production).

### Models

- **User** - User accounts
- **Trip** - Travel itineraries
- **Destination** - Locations within trips
- **DayPlan** - Daily activities
- **Activity** - Individual activities
- **PublicTrip** - Public trip metadata

### Migrations

Migrations are handled automatically via GORM AutoMigrate on startup.

## Testing

Run tests:

```bash
make test
```

Run with coverage:

```bash
make test-coverage
```

## Production Deployment

1. Set environment variables for production
2. Use PostgreSQL instead of SQLite
3. Set a secure `JWT_SECRET`
4. Configure Google OAuth credentials
5. Build and run:

```bash
make prod
```

## Contributing

1. Follow the existing project structure
2. Write tests for new features
3. Format code with `make fmt`
4. Run linter with `make lint`
5. Update documentation

## License

MIT License

# WordPress One-Click Docker Platform

A SaaS platform that enables users to deploy and manage WordPress websites with a single click using Docker on a single VPS.

## Architecture

- **Frontend**: Next.js 14 with TypeScript and Tailwind CSS
- **Backend**: Go API with Docker Engine orchestration
- **Database**: PostgreSQL for platform metadata
- **Reverse Proxy**: Traefik v3 with automatic Let's Encrypt SSL
- **Deployment**: Docker Compose with isolated WordPress stacks per project

## Features

### Must Have (MVP)
- Multi-tenant SaaS architecture with JWT authentication
- 1-click WordPress deployment per project
- Dynamic reverse proxy with Traefik v3
- Automatic Let's Encrypt SSL per domain (HTTP + DNS challenge)
- Cloudflare support (Full Strict mode)
- Custom domain mapping per project
- Start / Stop / Restart WordPress instances
- Container logs viewing
- Manual and scheduled automatic backups
- Project resource limits (CPU, memory)
- Basic monitoring (CPU, memory, disk usage)
- Environment variable management per project
- Safe project deletion with cleanup

### Project Structure

```
.
├── backend/              # Go backend API
│   ├── internal/         # Private application code
│   │   ├── api/         # HTTP handlers and routes
│   │   ├── auth/        # JWT authentication & RBAC
│   │   ├── docker/      # Docker client wrapper
│   │   ├── provision/   # WordPress provisioning logic
│   │   ├── backup/      # Backup and restore system
│   │   ├── monitor/     # Container stats collection
│   │   ├── domain/      # Domain validation & SSL
│   │   ├── scheduler/   # Cron-based backup scheduler
│   │   ├── models/      # Database models
│   │   └── config/      # Configuration management
│   ├── cmd/api/         # Application entry point
│   └── pkg/utils/       # Shared utilities
├── frontend/            # Next.js frontend
│   ├── src/
│   │   ├── app/        # Next.js 14 app router
│   │   ├── components/ # React components
│   │   ├── lib/        # Utility functions
│   │   └── types/      # TypeScript types
│   └── public/         # Static assets
├── traefik/            # Traefik configuration
└── backups/            # Backup storage directory
```

## Getting Started

### Prerequisites
- Docker and Docker Compose v2+
- Go 1.21+
- Node.js 20+
- PostgreSQL 14+
- A VPS with a public IP address
- Domain names pointing to your server's IP

### Environment Setup

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Configure the required environment variables:
```env
# Server configuration
SERVER_PORT=8080
SERVER_IP=your.server.ip.address

# Database
DATABASE_URL=postgres://user:password@localhost:5432/wplatform?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-key-change-in-production
JWT_EXPIRATION=24h

# Traefik / SSL
TRAEFIK_EMAIL=admin@example.com
CLOUDFLARE_API_TOKEN=optional-cloudflare-token
CLOUDFLARE_EMAIL=optional-cloudflare-email

# Backup
BACKUP_BASE_PATH=/var/backups/wordpress

# WordPress
WORDPRESS_IMAGE=wordpress:latest
```

### Deployment with Docker Compose

1. Create the Docker network for public access:
```bash
docker network create wp-public
```

2. Start the infrastructure:
```bash
docker-compose up -d traefik postgres
```

3. Set up the PostgreSQL database:
```bash
# Connect to PostgreSQL and create the database and user
docker exec -it wpplatform-postgres psql -U postgres
```
```sql
CREATE DATABASE wplatform;
CREATE USER wplatform WITH ENCRYPTED PASSWORD 'your-password';
GRANT ALL PRIVILEGES ON DATABASE wplatform TO wplatform;
\q
```

4. Run database migrations (the app will auto-migrate on startup)

5. Start the backend service:
```bash
docker-compose up -d backend
```

6. Start the frontend service:
```bash
docker-compose up -d frontend
```

### Development Setup

**Backend:**
```bash
cd backend
go mod download
go run cmd/api/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

### API Documentation

API documentation is available at `/swagger/index.html` when the backend is running.

### Creating Your First Admin User

Use the registration endpoint or direct database insertion to create an admin user:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"your-password","role":"admin"}'
```

### Troubleshooting

**Container fails to start:**
- Check Docker logs: `docker logs <container-name>`
- Verify the wp-public network exists: `docker network ls`
- Ensure ports 80 and 443 are not in use by other services

**SSL certificate issues:**
- Verify DNS A record points to your server IP
- Check Traefik logs: `docker logs traefik`
- For Cloudflare, ensure API token has correct permissions

**Database connection errors:**
- Verify PostgreSQL is running: `docker ps`
- Check connection string in .env
- Ensure database and user exist

### Scaling Considerations

- Each WordPress site runs in its own container stack
- Typical VPS can handle 10-20 WordPress sites
- Monitor disk usage for backups
- Consider offloading backups to S3-compatible storage

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/register` | POST | Register new user |
| `/auth/login` | POST | Login and get JWT |
| `/api/projects` | GET | List all projects |
| `/api/projects` | POST | Create new project |
| `/api/projects/{id}` | GET | Get project details |
| `/api/projects/{id}/start` | POST | Start project containers |
| `/api/projects/{id}/stop` | POST | Stop project containers |
| `/api/projects/{id}/logs` | GET | Get container logs |
| `/api/projects/{id}/stats` | GET | Get resource stats |
| `/api/projects/{id}/backups` | GET | List backups |
| `/api/backups` | POST | Create backup |
| `/api/dns/check` | POST | Check DNS propagation |

## License

Proprietary

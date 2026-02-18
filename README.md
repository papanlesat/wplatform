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
- Docker and Docker Compose
- Go 1.21+
- Node.js 20+
- PostgreSQL 14+

### Quick Start

1. Clone and configure environment
2. Start Traefik reverse proxy
3. Run backend API
4. Run Next.js frontend

See individual README files in `backend/` and `frontend/` for detailed instructions.

## API Documentation

API documentation will be available at `/api/docs` when the backend is running.

## License

Proprietary

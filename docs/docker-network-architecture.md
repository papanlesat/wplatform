# Docker Network Architecture

The WordPress Platform uses a multi-network Docker architecture to provide isolation and secure routing.

## Network Overview

### wp-public (Shared External Network)

**Purpose**: External access via Traefik reverse proxy

**Attached Containers**:
- Traefik reverse proxy
- All WordPress containers

**Characteristics**:
- Bridge network driver
- Allows Traefik to route traffic to WordPress containers
- Enables HTTP/HTTPS access from external internet
- Provides Let's Encrypt ACME challenge capability

**Creation**:
```bash
docker network create --driver bridge --attachable wp-public
```

Or use the provided script:
```bash
./scripts/create-network.sh
```

### project_{id}_net (Per-Project Internal Networks)

**Purpose**: Isolated communication between WordPress and its MySQL database

**Attached Containers**:
- WordPress container (for a specific project)
- MySQL container (for the same project)

**Characteristics**:
- Bridge network driver per project
- Isolates database access to only its WordPress instance
- Prevents cross-project database access
- Improves security through network segmentation

**Naming Convention**: `project_{project_uuid}_net`

**Example**: `project_550e8400-e29b-41d4-a716-446655440000_net`

## Network Topology

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                           │
└─────────────────────┬───────────────────────────────────┘
                      │ HTTP/HTTPS
                      ▼
              ┌───────────────┐
              │    Traefik    │
              │  (Reverse     │
              │    Proxy)     │
              └───────┬───────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
        ▼             ▼             ▼
   ┌─────────┐  ┌─────────┐  ┌─────────┐
   │    WP1  │  │    WP2  │  │    WP3  │
   └────┬────┘  └────┬────┘  └────┬────┘
        │             │             │
        │ wp-public   │ wp-public   │ wp-public
        │             │             │
   ┌────▼────┐  ┌────▼────┐  ┌────▼────┐
   │  MySQL1 │  │  MySQL2 │  │  MySQL3 │
   └─────────┘  └─────────┘  └─────────┘
        │             │             │
        │ project_net  │ project_net  │ project_net
```

## Container Network Attachments

### WordPress Containers
Each WordPress container is attached to **two networks**:

1. **wp-public**: For external HTTP/HTTPS traffic via Traefik
2. **project_{id}_net**: For MySQL database access

```go
// WordPress container network configuration
container.NetworkSettings.Networks = map[string]container.NetworkSettings{
    "wp-public": {NetworkID: wpPublicNetworkID},
    "project_" + projectID + "_net": {NetworkID: projectNetworkID},
}
```

### MySQL Containers
Each MySQL container is attached to **one network**:

1. **project_{id}_net**: Private database access only

MySQL is **not** attached to wp-public, ensuring databases are never directly accessible from the internet.

## Traefik Configuration

Traefik monitors the `wp-public` network for Docker containers with Traefik labels.

```yaml
# docker-compose.yml Traefik configuration
traefik:
  networks:
    - wp-public
  command:
    - "--providers.docker=true"
    - "--providers.docker.network=wp-public"
```

## WordPress Container Labels

WordPress containers attached to wp-public must include Traefik labels:

```go
labels := map[string]string{
    "traefik.enable": "true",
    "traefik.http.routers." + projectID + ".rule": "Host(`" + domain + "`)",
    "traefik.http.routers." + projectID + ".entrypoints": "web,websecure",
    "traefik.http.routers." + projectID + ".tls": "true",
    "traefik.http.routers." + projectID + ".tls.certresolver": "myresolver",
    "traefik.http.services." + projectID + ".loadbalancer.server.port": "80",
}
```

## Security Considerations

1. **Network Segmentation**: Databases are isolated on per-project networks
2. **No Direct DB Access**: MySQL containers never exposed to wp-public
3. **Traefik as Gateway**: All external traffic goes through Traefik
4. **Automatic SSL**: Traefik handles SSL termination for all domains
5. **Container Names**: Use UUIDs to prevent naming conflicts

## Network Management via Go Backend

The backend provides functions for network management:

```go
// Create project-specific isolated network
func CreateProjectNetwork(projectID uuid.UUID) error

// Attach container to wp-public network
func AttachContainerToPublicNetwork(containerID string) error

// Attach container to project internal network
func AttachContainerToProjectNetwork(containerID, projectID string) error

// Remove project network when deleting project
func RemoveProjectNetwork(projectID string) error
```

## Troubleshooting

### Check Network List
```bash
docker network ls
```

### Inspect Network
```bash
docker network inspect wp-public
docker network inspect project_{id}_net
```

### List Containers on Network
```bash
docker network inspect wp-public --format '{{json .Containers}}'
```

### Test Connectivity Between Containers
```bash
# From WordPress container
docker exec wp-{project-id} ping mysql-{project-id}

# Check network attachment
docker inspect wp-{project-id} --format '{{json .NetworkSettings.Networks}}'
```

### Network Conflict Resolution
If you encounter network conflicts:
```bash
# Remove conflicting network
docker network rm wp-public

# Recreate using script
./scripts/create-network.sh
```

## Production Considerations

1. **Subnet Configuration**: For large-scale deployments, consider custom subnet configuration
2. **IPv6 Support**: Enable if required for your use case
3. **Network MTU**: Adjust for VPN or custom network environments
4. **DNS Resolution**: Consider custom DNS for container-to-container resolution
5. **Monitoring**: Monitor network usage and potential bottlenecks

## Cleanup

When removing a project:
1. Stop and remove WordPress and MySQL containers
2. Remove project-specific network
3. wp-public network remains shared across all projects

```bash
docker network rm project_{project_id}_net
```

**Warning**: Never remove wp-public while projects are running.

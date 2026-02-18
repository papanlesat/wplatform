# Docker Daemon Configuration

This document describes how to configure the Docker daemon for API access by the WP Platform backend.

## Local Socket Access (Default)

For single-server deployments, the Docker Unix socket provides secure access by default.

### Configuration

The Go backend connects to Docker using the Unix socket:

```go
client, err := client.NewClientWithOpts(client.FromEnv)
```

### Permissions

The backend process needs to run with permission to access `/var/run/docker.sock`:

**Option 1: Add backend user to docker group**
```bash
sudo usermod -aG docker wpplatform
```

**Option 2: Run backend as root**
Not recommended for production but simplest for development.

### Docker Compose Mount

The docker-compose.yml mounts the socket into the PostgreSQL container for migrations:
```yaml
volumes:
  - /var/run/docker.sock:/var/run/docker.sock:ro
```

## TLS Configuration (Optional - For Remote Access)

If you need remote Docker API access (e.g., backend on separate server from Docker):

### Generate CA, Server Certificates

```bash
# Create directory for certificates
mkdir -p ~/.docker
cd ~/.docker

# Generate CA private key
openssl genrsa -aes256 -out ca-key.pem 4096

# Generate CA certificate
openssl req -new -x509 -days 365 -key ca-key.pem -sha256 -out ca.pem

# Generate server private key
openssl genrsa -out server-key.pem 4096

# Generate server certificate signing request (CSR)
openssl req -subj "/CN=$HOSTNAME" -sha256 -new -key server-key.pem -out server.csr

# Configure server certificate extensions
echo subjectAltName = DNS:$HOSTNAME,IP:127.0.0.1 >> extfile.cnf
echo extendedKeyUsage = serverAuth >> extfile.cnf

# Sign the server certificate
openssl x509 -req -days 365 -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile extfile.cnf
```

### Generate Client Certificates

```bash
# Generate client private key
openssl genrsa -out client-key.pem 4096

# Generate client certificate signing request
openssl req -subj '/CN=client' -new -key client-key.pem -out client.csr

# Configure client certificate extensions
echo extendedKeyUsage = clientAuth > extfile-client.cnf

# Sign the client certificate
openssl x509 -req -days 365 -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile extfile-client.cnf
```

### Configure Docker Daemon

Edit `/etc/docker/daemon.json`:
```json
{
  "hosts": ["unix:///var/run/docker.sock", "tcp://0.0.0.0:2376"],
  "tls": true,
  "tlsverify": true,
  "tlscacert": "/root/.docker/ca.pem",
  "tlscert": "/root/.docker/server-cert.pem",
  "tlskey": "/root/.docker/server-key.pem"
}
```

Restart Docker:
```bash
sudo systemctl restart docker
```

### Connect with TLS from Go Backend

Set environment variables:
```bash
export DOCKER_HOST=tcp://your-docker-host:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/path/to/certificates
```

The Docker Go client will automatically use these certificates when initialized with `client.FromEnv`.

## Security Best Practices

1. **Unix Socket for Local**: Use Unix socket when backend and Docker are on the same server
2. **TLS for Remote**: Always use TLS verification for remote Docker API access
3. **Least Privilege**: Only give Docker socket access to necessary users
4. **Read-Only When Possible**: Mount sockets as read-only (`:ro`) where applicable
5. **Network Isolation**: Restrict which IPs can connect to Docker API port 2376

## Troubleshooting

### Permission Denied

```bash
# Check current user groups
groups

# Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in for changes to take effect
```

### Connection Refused

```bash
# Check if Docker daemon is running
sudo systemctl status docker

# Check Docker API port (if using remote)
netstat -tuln | grep 2376
```

### Certificate Issues

```bash
# Verify certificate contents
openssl x509 -in server-cert.pem -text -noout

# Verify server certificate matches CA
openssl verify -CAfile ca.pem server-cert.pem
```

## Production Deployment

For production:
1. Use Unix socket access (backend and Docker on same server)
2. Add backend user to docker group
3. Restrict Docker daemon to listen only on Unix socket
4. Monitor Docker API access logs
5. Use firewall to restrict network access

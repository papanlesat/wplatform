# Traefik v3 Configuration
#
# This directory contains Traefik configuration for the WordPress Platform
#
# Features:
# - HTTP challenge for Let's Encrypt (default)
# - DNS challenge for Cloudflare (optional)
# - Automatic SSL certificate management
# - Dynamic Docker provider integration
# - Dashboard at http://localhost:8080
#
# Network: wp-public (external network created by script)

## Files

- `docker-compose.yml` - Main Traefik service definition (in project root)
- `certs/` - Let's Encrypt certificate storage
- `traefik.yml` - Optional static configuration (can be used for DNS challenge)

## Configuration

### Default Setup (HTTP Challenge)

The docker-compose.yml in the project root configures Traefik with:

1. **HTTP Challenge** - Default for most domains
   - Challenge type: TLS-ALPN-01
   - Challenge port: 80
   - Storage: `./certs/acme.json`

2. **Entry Points**
   - `web` - HTTP (port 80)
   - `websecure` - HTTPS (port 443)

3. **Certificate Resolver** - `letsencrypt`
   - Email: Configured via `TRAEFIK_EMAIL` environment variable
   - CA Server: Let's Encrypt v2

4. **Docker Provider**
   - Monitors Docker for containers with Traefik labels
   - Network: `wp-public` only
   - Exposed by default: false (containers must opt-in)

### Cloudflare DNS Challenge (Optional)

To enable DNS challenge for Cloudflare Full Strict mode:

1. Set environment variables in `.env`:
   ```
   CLOUDFLARE_EMAIL=your-cloudflare-email@example.com
   CLOUDFLARE_API_TOKEN=your_api_token_here
   ```

2. Create or modify `traefik/traefik.yml`:
   ```yaml
   certificatesResolvers:
     cloudflare:
       acme:
         email: your-email@example.com
         storage: /letsencrypt/acme-cloudflare.json
         dnsChallenge:
           provider: cloudflare
           delayBeforeCheck: 0
           resolvers:
             - "1.1.1.1"
           - "1.0.0.1"
   ```

3. Update docker-compose.yml to include additional DNS challenge commands:
   ```yaml
   command:
     - "--certificatesresolvers.cloudflare.acme.dnschallenge=true"
     - "--certificatesresolvers.cloudflare.acme.dnschallenge.provider=cloudflare"
   ```

## Docker Network Requirements

Traefik requires the `wp-public` network to exist:

```bash
# Create network if it doesn't exist
docker network create wp-public || true
```

WordPress containers will:
- Attach to `wp-public` for external access
- Attach to `project_{id}_net` for internal database communication
- Include Traefik labels to enable routing

## Container Labels

WordPress containers must include these labels to be routed by Traefik:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.{project}.rule=Host(`example.com`)"
  - "traefik.http.routers.{project}.entrypoints=web,websecure"
  - "traefik.http.routers.{project}.tls=true"
  - "traefik.http.routers.{project}.tls.certresolver=letsencrypt"
  - "traefik.http.services.{project}.loadbalancer.server.port=80"
```

## Access Points

- **Dashboard**: http://localhost:8080/dashboard/
- **API**: http://localhost:8080/api/
- **Metrics**: Not enabled in this configuration

## SSL/TLS Configuration

### HTTP Challenge (Default)
- Validates domain ownership via HTTP
- Works for all standard domains
- Automatic certificate renewal

### DNS Challenge (Cloudflare)
- Faster than HTTP challenge
- Required for wildcard certificates
- Works with Cloudflare Full Strict mode
- No need to expose port 80 for validation

### Certificate Storage

Certificates are stored in `./certs/` directory:

- `acme.json` - HTTP challenge certificates
- `acme-cloudflare.json` - DNS challenge certificates (if enabled)

**Security**: Ensure `certs/` directory is backed up regularly

## Troubleshooting

### Traefik won't start
- Check if port 80, 443, 8080 are available
- Check if Docker daemon is running
- Review logs: `docker logs traefik`

### Certificate issuance failing
- Verify domain DNS points to this server
- Check firewall allows ports 80 and 443
- Review Traefik dashboard for ACME errors
- Check Let's Encrypt rate limits

### Containers not being routed
- Verify container is on `wp-public` network: `docker network inspect wp-public`
- Check container labels are correct
- Ensure `traefik.enable=true` is present
- Check Traefik dashboard for provider errors

## Production Deployment

For production deployment:

1. Update `TRAEFIK_EMAIL` to a real email address
2. Enable SSL/TLS in production (remove `--api.insecure=true`)
3. Configure proper firewall rules
4. Set up automated backups for `certs/` directory
5. Monitor Traefik logs for errors
6. Set up Cloudflare if using DNS challenge

## Security Considerations

1. **Dashboard Access**: Remove `--api.insecure=true` in production
2. **Certificate Storage**: Restrict access to `certs/` directory
3. **Environment Variables**: Store Cloudflare tokens securely
4. **Network Isolation**: MySQL containers on private networks only
5. **Rate Limiting**: Let's Encrypt has production rate limits

## Rate Limits (Let's Encrypt)

As of 2024, Let's Encrypt rate limits:

- **Certificates per Registered Domain**: 50 per week
- **Duplicate Certificates**: 5 per week
- **Failed Validations**: 5 failures per account, per hostname, per hour
- **New Orders**: 10 orders per 3 hours

**Best Practices**:
- Test staging environment with Let's Encrypt staging endpoint
- Implement caching to reduce certificate requests
- Use DNS challenge when possible
- Monitor certificate expiration dates

## Backup Strategy

Back up the `certs/` directory regularly:

```bash
# Backup certificates
tar -czf traefik-certs-backup-$(date +%Y%m%d).tar.gz traefik/

# Store backups in secure, offsite location
# Keep at least 30 days of backups
```

## Monitoring

Monitor Traefik health:

```bash
# Check Traefik logs
docker logs -f traefik

# Check Traefik dashboard
curl http://localhost:8080/api/rawdata

# Check certificate status
docker exec traefik ls -la /letsencrypt/
```

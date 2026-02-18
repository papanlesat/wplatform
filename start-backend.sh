#!/bin/bash
export DATABASE_URL="postgres://wpplatform_user:wpplatform_pass@localhost:5432/wpplatform_db?sslmode=disable"
export JWT_SECRET="dev-secret-key"
export JWT_EXPIRATION="24h"
export SERVER_PORT="8080"
export SERVER_IP="127.0.0.1"
cd /home/andri/Documents/Docker/wpdocker-platform/backend
exec /tmp/backend

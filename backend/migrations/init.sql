
CREATE TYPE user_role AS ENUM ('admin', 'user');

CREATE TYPE project_status AS ENUM ('creating', 'running', 'stopped', 'error');

CREATE TYPE backup_type AS ENUM ('manual', 'scheduled');

CREATE TYPE ssl_status AS ENUM ('pending', 'active', 'failed');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    status project_status NOT NULL DEFAULT 'creating',
    cpu_limit FLOAT NOT NULL DEFAULT 0.5,
    memory_limit_mb INTEGER NOT NULL DEFAULT 512,
    wordpress_container_name VARCHAR(255),
    mysql_container_name VARCHAR(255),
    db_name VARCHAR(255) NOT NULL,
    db_user VARCHAR(255) NOT NULL,
    db_password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_domain ON projects(domain);

CREATE TABLE backups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    file_path VARCHAR(1024) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    type backup_type NOT NULL DEFAULT 'manual',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_backups_project_id ON backups(project_id);
CREATE INDEX idx_backups_type ON backups(type);
CREATE INDEX idx_backups_created_at ON backups(created_at);

CREATE TABLE domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    domain_name VARCHAR(255) UNIQUE NOT NULL,
    ssl_status ssl_status NOT NULL DEFAULT 'pending',
    cloudflare_api_token VARCHAR(512),
    cloudflare_email VARCHAR(255),
    dns_challenge_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_domains_project_id ON domains(project_id);
CREATE INDEX idx_domains_domain_name ON domains(domain_name);
CREATE INDEX idx_domains_ssl_status ON domains(ssl_status);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_domains_updated_at
    BEFORE UPDATE ON domains
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE environment_variables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id, key)
);

CREATE INDEX idx_env_vars_project_id ON environment_variables(project_id);

CREATE TRIGGER update_environment_variables_updated_at
    BEFORE UPDATE ON environment_variables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE backup_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    schedule_cron VARCHAR(255) NOT NULL DEFAULT '0 2 * * *',
    retention_days INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(project_id)
);

CREATE INDEX idx_backup_schedules_project_id ON backup_schedules(project_id);

CREATE TRIGGER update_backup_schedules_updated_at
    BEFORE UPDATE ON backup_schedules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO users (email, password_hash, role) VALUES
('admin@wp-platform.com', '$2a$10$dummy.hash.change.in.production', 'admin');

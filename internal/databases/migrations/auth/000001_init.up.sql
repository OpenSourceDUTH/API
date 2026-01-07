-- Define (default) quota tiers for user groups.
-- For our usecases the groups will be manually assigned quotas at the end of this script.
-- Default values concern mainly possible future groups.
CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    default_rpm INTEGER NOT NULL DEFAULT 60,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Domains that automatically grant academic group membership
CREATE TABLE academic_domains (
    domain TEXT PRIMARY KEY
);

-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    group_id INTEGER NOT NULL,
    max_tokens INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id)
);

-- OAuth identities linking users to providers
CREATE TABLE oauth_identities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN ('google', 'github')),
    provider_id TEXT NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE (user_id, provider),
    UNIQUE (provider, provider_id)
);

-- OAuth states for CSRF protection
CREATE TABLE oauth_states (
    state TEXT PRIMARY KEY,
    expires_at TIMESTAMP NOT NULL
);

-- Server-side sessions
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- API features (hierarchical)
CREATE TABLE features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    parent_id INTEGER,
    admin_only BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES features(id) ON DELETE SET NULL
);

-- Group-level quota defaults per feature
CREATE TABLE group_feature_quotas (
    group_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    rpm_limit INTEGER, -- NULL means uncapped
    PRIMARY KEY (group_id, feature_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE
);

-- Per-user quota overrides (takes precedence over group)
CREATE TABLE user_quota_overrides (
    user_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    rpm_limit INTEGER, -- NULL means uncapped
    PRIMARY KEY (user_id, feature_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE
);

-- API tokens (opaque bearer tokens)
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    admin_created BOOLEAN NOT NULL DEFAULT 0,
    expires_at TIMESTAMP, -- NULL means non-expiring
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Token feature scopes
CREATE TABLE token_features (
    token_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    PRIMARY KEY (token_id, feature_id),
    FOREIGN KEY (token_id) REFERENCES tokens(id) ON DELETE CASCADE,
    FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE
);

-- Token IP whitelist (canonicalized IPs)
CREATE TABLE token_allowed_ips (
    token_id INTEGER NOT NULL,
    ip_address TEXT NOT NULL,
    PRIMARY KEY (token_id, ip_address),
    FOREIGN KEY (token_id) REFERENCES tokens(id) ON DELETE CASCADE
);

-- Usage log for rate limiting
CREATE TABLE usage_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (feature_id) REFERENCES features(id) ON DELETE CASCADE
);

-- Index for efficient RPM queries
CREATE INDEX idx_usage_log_rpm ON usage_log(user_id, feature_id, timestamp);

-- Index for session lookups
CREATE INDEX idx_sessions_user ON sessions(user_id);

-- Index for token lookups
CREATE INDEX idx_tokens_user ON tokens(user_id);

-- Seed default groups
INSERT INTO groups (name, default_rpm, description) VALUES 
    ('regular', 60, 'Default group for regular users'),
    ('academic', 120, 'Academic users with higher quotas');

-- Seed academic domains
INSERT INTO academic_domains (domain) VALUES ('cs.duth.gr');

-- Seed core features
INSERT INTO features (slug, name, parent_id, admin_only) VALUES 
    ('maps', 'Maps API', NULL, 0),
    ('schedule', 'Schedule API', NULL, 0),
    ('search', 'Search API', NULL, 0);

-- Seed maps sub-features
INSERT INTO features (slug, name, parent_id, admin_only) VALUES 
    ('maps.tiles', 'Map Tiles', (SELECT id FROM features WHERE slug = 'maps'), 0),
    ('maps.routing', 'Routing API', (SELECT id FROM features WHERE slug = 'maps'), 0),
    ('maps.geocoding', 'Geocoding API', (SELECT id FROM features WHERE slug = 'maps'), 0);

-- Seed default group quotas for features
INSERT INTO group_feature_quotas (group_id, feature_id, rpm_limit)
SELECT g.id, f.id, 
    CASE 
        WHEN f.slug = 'maps.tiles' THEN CASE WHEN g.name = 'academic' THEN 500 ELSE 250 END
        WHEN f.slug = 'maps.routing' THEN CASE WHEN g.name = 'academic' THEN 100 ELSE 50 END
        WHEN f.slug = 'maps.geocoding' THEN CASE WHEN g.name = 'academic' THEN 100 ELSE 50 END
        WHEN f.slug = 'maps' THEN CASE WHEN g.name = 'academic' THEN 200 ELSE 100 END
        WHEN f.slug = 'schedule' THEN CASE WHEN g.name = 'academic' THEN 120 ELSE 60 END
        WHEN f.slug = 'search' THEN CASE WHEN g.name = 'academic' THEN 120 ELSE 60 END
        ELSE g.default_rpm
    END
FROM groups g
CROSS JOIN features f;


-- This project is the monolithic backend API for the OpenSourceDUTH team. Access to open data compiled and provided by the OpenSourceDUTH University Team.
-- API Copyright (C) 2025 OpenSourceDUTH
--     This program is free software: you can redistribute it and/or modify
--     it under the terms of the GNU General Public License as published by
--     the Free Software Foundation, either version 3 of the License, or
--     (at your option) any later version.

--     This program is distributed in the hope that it will be useful,
--     but WITHOUT ANY WARRANTY; without even the implied warranty of
--     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
--     GNU General Public License for more details.

--     You should have received a copy of the GNU General Public License
--     along with this program.  If not, see <https://www.gnu.org/licenses/>.
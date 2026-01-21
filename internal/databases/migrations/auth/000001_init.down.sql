-- Drop indexes
DROP INDEX IF EXISTS idx_usage_log_rpm;
DROP INDEX IF EXISTS idx_sessions_user;
DROP INDEX IF EXISTS idx_tokens_user;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS usage_log;
DROP TABLE IF EXISTS token_allowed_ips;
DROP TABLE IF EXISTS token_features;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS user_quota_overrides;
DROP TABLE IF EXISTS group_feature_quotas;
DROP TABLE IF EXISTS features;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS oauth_identities;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS academic_domains;
DROP TABLE IF EXISTS groups;

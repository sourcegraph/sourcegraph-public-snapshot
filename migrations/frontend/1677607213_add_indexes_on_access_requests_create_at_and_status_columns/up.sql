CREATE INDEX IF NOT EXISTS access_requests_created_at ON access_requests (created_at);
CREATE INDEX IF NOT EXISTS access_requests_status ON access_requests (status);
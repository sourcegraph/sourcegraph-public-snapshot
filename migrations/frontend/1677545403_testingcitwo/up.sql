-- Perform migration here.
CREATE TABLE IF NOT EXISTS testci_two(
    id UUID NOT NULL,
    test TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
    );

COMMENT ON TABLE testci_two IS 'This is just a test to see if CI would catch migraion';
COMMENT ON COLUMN testci_two.test IS 'just for test';

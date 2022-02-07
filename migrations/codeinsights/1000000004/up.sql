-- Disables TimescaleDB telemetry, which we cannot easily ship with Sourcegraph
-- in a reasonable way (requires fairly in-depth analysis of what gets sent, etc.)
-- See https://docs.timescale.com/latest/using-timescaledb/telemetry
--
-- Cannot be run inside of a transaction block.
ALTER SYSTEM SET timescaledb.telemetry_level=off;

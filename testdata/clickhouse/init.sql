CREATE DATABASE IF NOT EXISTS audit_test;

-- Suboptimal schema: high-cardinality first in ORDER BY, String timestamp, status without LowCardinality
CREATE TABLE IF NOT EXISTS audit_test.events_bad
(
    user_uuid UUID,
    day Date,
    created_at String,
    status String,
    value Float64
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(day)
ORDER BY (user_uuid, day);

INSERT INTO audit_test.events_bad
SELECT
    generateUUIDv4(),
    toDate('2024-01-01') + (number % 28),
    toString(now()),
    ['ok', 'fail', 'pending'][1 + (number % 3)],
    rand() / 1000
FROM numbers(500);

-- Better reference table
CREATE TABLE IF NOT EXISTS audit_test.events_ok
(
    day Date,
    status LowCardinality(String),
    user_id UInt64,
    created_at DateTime64(3),
    value Float64
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(day)
ORDER BY (day, status, user_id);

INSERT INTO audit_test.events_ok
SELECT
    toDate('2024-02-01') + (number % 14),
    ['ok', 'fail'][1 + (number % 2)],
    number,
    now64(3),
    rand() / 100
FROM numbers(200);

-- MV chain
CREATE TABLE IF NOT EXISTS audit_test.events_daily
(
    day Date,
    status String,
    cnt UInt64
)
ENGINE = SummingMergeTree
ORDER BY (day, status);

CREATE MATERIALIZED VIEW IF NOT EXISTS audit_test.events_bad_mv
TO audit_test.events_daily
AS
SELECT day, status, count() AS cnt
FROM audit_test.events_bad
GROUP BY day, status;

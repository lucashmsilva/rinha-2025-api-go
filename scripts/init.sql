SELECT 'CREATE DATABASE rinha_pay'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'rinha_pay')\gexec

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL NOT NULL,
    correlation_id UUID NULL DEFAULT NULL,
    amount BIGINT NULL DEFAULT NULL,
    processor_used VARCHAR(8) NULL DEFAULT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT payments_correlation_id_key UNIQUE (correlation_id)
);

CREATE INDEX requested_at_idx ON payments (requested_at);

CREATE TABLE IF NOT EXISTS processor_health (
    processor VARCHAR(8) NOT NULL,
    is_falling BOOLEAN NULL DEFAULT NULL,
    min_response_time BIGINT NULL DEFAULT NULL,
    falling_cycles BIGINT NULL DEFAULT NULL,
    PRIMARY KEY (processor)
);

INSERT INTO processor_health
(processor, is_falling, min_response_time, falling_cycles)
VALUES ('default', 'false', 0, 0);

INSERT INTO processor_health
(processor, is_falling, min_response_time, falling_cycles)
VALUES ('fallback', 'false', 0, 0);

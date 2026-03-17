-- FSC CMS POC - Initial schema and seed data

CREATE TABLE IF NOT EXISTS stops (
    id   VARCHAR(32) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    lat  DOUBLE PRECISION NOT NULL,
    lng  DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS vehicles (
    id              VARCHAR(32) PRIMARY KEY,
    line            VARCHAR(32) NOT NULL,
    current_stop_id VARCHAR(32) NOT NULL REFERENCES stops(id),
    lat             DOUBLE PRECISION NOT NULL,
    lng             DOUBLE PRECISION NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS passengers (
    card_id   VARCHAR(64) PRIMARY KEY,
    name      VARCHAR(128) NOT NULL,
    category  VARCHAR(32)  NOT NULL DEFAULT 'regular',
    is_active BOOLEAN      NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS validation_events (
    id         BIGSERIAL PRIMARY KEY,
    card_id    VARCHAR(64)  NOT NULL REFERENCES passengers(card_id),
    vehicle_id VARCHAR(32)  NOT NULL REFERENCES vehicles(id),
    event_type VARCHAR(16)  NOT NULL CHECK (event_type IN ('checkin', 'checkout')),
    stop_id    VARCHAR(32)  NOT NULL REFERENCES stops(id),
    lat        DOUBLE PRECISION NOT NULL,
    lng        DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_validation_card_type_time
    ON validation_events (card_id, event_type, created_at DESC);

CREATE INDEX idx_validation_created_at
    ON validation_events (created_at DESC);

-- OD Matrix: pair each checkin with the chronologically next checkout
-- for the same card_id, then aggregate by origin + destination stop.
CREATE MATERIALIZED VIEW od_matrix AS
SELECT
    ci.stop_id   AS origin_stop,
    os.name      AS origin_name,
    co.stop_id   AS destination_stop,
    ds.name      AS destination_name,
    COUNT(*)     AS trip_count
FROM validation_events ci
JOIN LATERAL (
    SELECT ve.stop_id, ve.created_at
    FROM validation_events ve
    WHERE ve.card_id    = ci.card_id
      AND ve.event_type = 'checkout'
      AND ve.created_at > ci.created_at
    ORDER BY ve.created_at ASC
    LIMIT 1
) co ON TRUE
JOIN stops os ON os.id = ci.stop_id
JOIN stops ds ON ds.id = co.stop_id
WHERE ci.event_type = 'checkin'
GROUP BY ci.stop_id, os.name, co.stop_id, ds.name;

CREATE UNIQUE INDEX idx_od_matrix_pair ON od_matrix (origin_stop, destination_stop);

-- Seed data: Bucharest line 41 stops
INSERT INTO stops (id, name, lat, lng) VALUES
    ('S1', 'Piața Unirii',    44.4268, 26.1025),
    ('S2', 'Universitate',    44.4361, 26.1006),
    ('S3', 'Piața Romană',    44.4478, 26.0934),
    ('S4', 'Aviatorilor',     44.4563, 26.0877),
    ('S5', 'Aurel Vlaicu',    44.4634, 26.0912),
    ('S6', 'Piața Victoriei', 44.4519, 26.0793);

-- Seed data: 3 vehicles on line 41
INSERT INTO vehicles (id, line, current_stop_id, lat, lng) VALUES
    ('BUS-101', '41', 'S1', 44.4268, 26.1025),
    ('BUS-102', '41', 'S3', 44.4478, 26.0934),
    ('BUS-103', '41', 'S5', 44.4634, 26.0912);

-- Seed data: passengers — subsidized and regular
INSERT INTO passengers (card_id, name, category, is_active) VALUES
    ('CMS-001', 'Ion Popescu',      'student',    TRUE),
    ('CMS-002', 'Maria Ionescu',    'university', TRUE),
    ('CMS-003', 'Gheorghe Marin',   'pensioner',  TRUE),
    ('CMS-004', 'Elena Dumitrescu', 'disabled',   TRUE),
    ('CMS-005', 'Vasile Stanescu',  'veteran',    TRUE),
    ('CMS-006', 'Andrei Radu',      'regular',    TRUE),
    ('CMS-007', 'Cristina Popa',    'regular',    TRUE),
    ('CMS-008', 'Mihai Georgescu',  'regular',    TRUE);

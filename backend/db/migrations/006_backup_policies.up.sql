CREATE TABLE backup_policies (
    id             BIGSERIAL PRIMARY KEY,
    client_id      BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    destination    TEXT NOT NULL,
    retain_last    INT NOT NULL DEFAULT 0,
    retain_hourly  INT NOT NULL DEFAULT 0,
    retain_daily   INT NOT NULL DEFAULT 7,
    retain_weekly  INT NOT NULL DEFAULT 4,
    retain_monthly INT NOT NULL DEFAULT 12,
    retain_yearly  INT NOT NULL DEFAULT 3,
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE backup_schedule_times (
    id        BIGSERIAL PRIMARY KEY,
    policy_id BIGINT NOT NULL REFERENCES backup_policies(id) ON DELETE CASCADE,
    run_at    TIME NOT NULL,
    UNIQUE(policy_id, run_at)
);

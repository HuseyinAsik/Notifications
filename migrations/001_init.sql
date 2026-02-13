-- UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =========================
-- NOTIFICATIONS TABLE
-- =========================

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    group_id UUID NOT NULL,
    recipient TEXT NOT NULL,
    channel VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    priority VARCHAR(20) NOT NULL,

    scheduled_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- -------------------------
-- INDEXES FOR FILTERING
-- -------------------------

-- Status filter
CREATE INDEX IF NOT EXISTS idx_notifications_status
ON notifications (status);

-- Channel filter
CREATE INDEX IF NOT EXISTS idx_notifications_channel
ON notifications (channel);

-- Date range + pagination
CREATE INDEX IF NOT EXISTS idx_notifications_created_at_desc
ON notifications (created_at DESC);

-- Combined filtering (status + channel + created_at)
CREATE INDEX IF NOT EXISTS idx_notifications_filtering
ON notifications (status, channel, created_at DESC);

-- Scheduled jobs query (worker için)
CREATE INDEX IF NOT EXISTS idx_notifications_scheduled
ON notifications (status, scheduled_at)
WHERE scheduled_at IS NOT NULL;



-- =========================
-- OUTBOX EVENTS TABLE
-- =========================

CREATE TABLE IF NOT EXISTS outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    aggregate_id UUID NOT NULL,
    group_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    topic VARCHAR(100) NOT NULL,
    payload BYTEA NOT NULL,

    status VARCHAR(20) NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP NULL
);

-- -------------------------
-- OUTBOX INDEXES
-- -------------------------

-- Worker polling (pending events)
CREATE INDEX IF NOT EXISTS idx_outbox_status_created
ON outbox_events (status, created_at);

-- Aggregate lookup
CREATE INDEX IF NOT EXISTS idx_outbox_aggregate_id
ON outbox_events (aggregate_id);

-- Retry logic
CREATE INDEX IF NOT EXISTS idx_outbox_retry
ON outbox_events (status, retry_count);

-- Published cleanup queries
CREATE INDEX IF NOT EXISTS idx_outbox_published_at
ON outbox_events (published_at);

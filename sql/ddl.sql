CREATE TABLE IF NOT EXISTS PAYMENTS(
    id uuid not null primary key,
    amount numeric not null,
    payment_processor varchar(20) not null,
    requested_at TIMESTAMPTZ not null
)
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS events;

-- =====================
-- EVENTS
-- =====================
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    total_seats INT NOT NULL,
    ends_at TIMESTAMP NOT NULL
);

-- =====================
-- RESERVATIONS
-- =====================
CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(id),

    name TEXT NOT NULL,
    email TEXT NOT NULL,

    quantity INT NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',
    token TEXT NOT NULL,

    created_at TIMESTAMP DEFAULT NOW()
);
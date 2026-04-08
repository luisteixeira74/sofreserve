CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    total_seats INT NOT NULL
);

CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL,
    name TEXT,
    quantity INT NOT NULL
);
CREATE TABLE reservation_tickets (
    id BIGSERIAL PRIMARY KEY,

    reservation_id BIGINT NOT NULL,

    ticket_number INT NOT NULL,

    token VARCHAR(255) NOT NULL UNIQUE,

    checked_in_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_reservation_tickets_reservation
        FOREIGN KEY (reservation_id)
        REFERENCES reservations(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_reservation_ticket_number
        UNIQUE (reservation_id, ticket_number)
);

CREATE INDEX idx_reservation_tickets_reservation_id
    ON reservation_tickets(reservation_id);

CREATE INDEX idx_reservation_tickets_token
    ON reservation_tickets(token);
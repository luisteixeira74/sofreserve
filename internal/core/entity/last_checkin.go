package entity

import "time"

type LastCheckin struct {
    TicketNumber int
    Token        string
    CheckedInAt  time.Time
    CheckedInAtStr string
}
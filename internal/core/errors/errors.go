package errors

import "errors"

var (
    ErrNotEnoughSeats   = errors.New("not enough seats")
    ErrEventNotFound    = errors.New("event not found")
    ErrInvalidEventID   = errors.New("invalid event id")
    ErrInvalidQuantity  = errors.New("invalid quantity")
    ErrInvalidName      = errors.New("invalid name")
)
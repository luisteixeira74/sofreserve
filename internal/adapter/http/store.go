package http

import "sof-reserve/internal/core/entity"

var events = make(map[int]entity.Event)
var nextID = 1
var reservations = []entity.Reservation{}
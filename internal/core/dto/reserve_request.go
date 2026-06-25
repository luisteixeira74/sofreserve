package dto

type ReserveRequest struct {
	EventID  int64
	EventName string
	Name     string
	Email    string
	Quantity int
}
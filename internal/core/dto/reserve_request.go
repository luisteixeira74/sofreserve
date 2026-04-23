package dto

type ReserveRequest struct {
	EventID  int
	EventName string
	Name     string
	Email    string
	Quantity int
}
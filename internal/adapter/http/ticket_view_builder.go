package http

import (
	"sof-reserve/internal/core/entity"
	"sof-reserve/internal/shared/message"
)

type TicketViewBuilder struct{}

func (b TicketViewBuilder) Build(
	baseURL string,
	eventName string,
	t ReservationTicketView,
) ReservationTicketView {

	t.DisplayToken = t.Token

	t.QRCodeURL = baseURL + "/ticket/" + t.Token

	t.WhatsAppLink = message.BuildTicketWhatsAppMessage(
		baseURL,
		entity.TicketView{
			EventName:    eventName,
			Token:        t.Token,
			TicketNumber: t.Number,
		},
	)

	return t
}
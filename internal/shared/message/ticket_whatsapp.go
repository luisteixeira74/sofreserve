package message

import (
	"fmt"
	"net/url"

	"sof-reserve/internal/core/entity"
)

func BuildTicketWhatsAppMessage(
	baseURL string,
	ticket entity.TicketView,
) string {

	ticketURL := fmt.Sprintf(
		"%s/ticket/%s",
		baseURL,
		ticket.Token,
	)

	message := fmt.Sprintf(
`SOFRESERVE • Ticket Individual

Evento:
%s

Ticket:
#%d

Acesso do participante:
%s

Importante:
Cada participante possui um ticket individual.
O check-in é realizado separadamente.

—
Powered by SOFRESERVE
%s`,
		ticket.EventName,
		ticket.TicketNumber,
		ticketURL,
		baseURL,
	)

	return "https://wa.me/?text=" + url.QueryEscape(message)
}
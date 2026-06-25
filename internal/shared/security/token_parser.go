package security

import "strings"

func NormalizeTicketToken(token string) string {
	return strings.TrimPrefix(token, TicketPrefix)
}
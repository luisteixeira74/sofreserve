package security

func GenerateReservationToken() (string, error) {
	return generateRandomToken(16)
}
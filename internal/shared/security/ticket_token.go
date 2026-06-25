package security

func GenerateTicketToken() (string, error) {
	token, err := generateRandomToken(16)
	if err != nil {
		return "", err
	}

	return TicketPrefix + token, nil
}
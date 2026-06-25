package security

func GenerateOwnerToken() (string, error) {
	return generateRandomToken(16)
}
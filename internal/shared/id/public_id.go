package id

import (
	"strconv"
	"strings"
	"time"
)

func GeneratePublicID() string {
	id := strings.ToLower(strconv.FormatInt(time.Now().UnixNano(), 36))

	if len(id) > 8 {
		return id[:8]
	}
	return id
}
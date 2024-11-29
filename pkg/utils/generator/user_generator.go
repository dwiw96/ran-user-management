package generator

import (
	"math/rand"
	"strings"
)

const alphabet string = "abcdefghijklmnopqrstuvwxyz"

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandomInt32(min, max int32) int32 {
	return min + rand.Int31n(max-min+1)
}

func CreateRandomString(length int) string {
	var res strings.Builder

	alphabetLen := len(alphabet)
	for i := 0; i < length; i++ {
		char := alphabet[rand.Intn(alphabetLen)]
		res.WriteByte(char)
	}

	return res.String()
}

func CreateRandomEmail(name string) string {
	return name + "@mail.com"
}

package utils

import "strings"

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Encode(number uint64) string {
	length := uint64(len(alphabet))
	var encodedBuilder strings.Builder

	if number == 0 {
		return string(alphabet[0])
	}

	for number > 0 {
		encodedBuilder.WriteByte(alphabet[number%length])
		number = number / length
	}

	return reverse(encodedBuilder.String())
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

package main

import (
	"strings"
	"math/rand"
)

func splitString(str string) []string {
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ".")
}

func makeID(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
			letterIdxBits = 6                    // 6 bits to represent a letter index
			letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
			letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
			if remain == 0 {
					cache, remain = rand.Int63(), letterIdxMax
			}
			if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
					b[i] = letterBytes[idx]
					i--
			}
			cache >>= letterIdxBits
			remain--
	}

	return string(b)
}


// refer https://yourbasic.org/golang/delete-element-slice/
func removeItemFromSlice (clients []*Client, c *Client) {
	for index, client := range clients {
		if client.id == c.id {
			clients[index] = clients[len(clients)-1] // Copy last element to index i.
			clients[len(clients)-1] = nil   // Erase last element (write zero value).
			clients = clients[:len(clients)-1]   // Truncate slice.
		}
	}
}
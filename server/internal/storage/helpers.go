package storage

import (
	"crypto/rand"
	"fmt"
)

func newUuid() string {
	buff := make([]byte, 16)
	_, err := rand.Read(buff)
	if err != nil {
		panic(err)
	}

	buff[6] = (buff[6] & 0x0f) | 0x40
	buff[8] = (buff[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		buff[0:4], buff[4:6], buff[6:8], buff[8:10], buff[10:])
}

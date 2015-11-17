package main

import (
	"bufio"
	"image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	ln, _ := net.Listen("tcp", ":29300")
	conn, _ := ln.Accept()

	for {
		reader := bufio.NewReader(conn)
		message, _ := reader.ReadString('\n')
		log.Println("Message Received:", string(message))

		conn.Write([]byte("thing"))

		lenRune, _, err := reader.ReadRune()
		len := int32(lenRune)
        data := [len]byte

		f, err := os.Create("/var/www/image/" + imageName() + ".png")
		if err != nil {
			panic(err)
		}
		err = png.Encode(f, img)
		if err != nil {
			panic(err)
		}
		f.Close()
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func imageName() string {
	b := make([]byte, 5)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := 4, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
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

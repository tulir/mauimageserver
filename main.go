package main

import (
	/*"bufio"
	"log"*/
	flag "github.com/ogier/pflag"
	"image/png"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	dirPtr := flag.StringP("directory", "d", "./", "The directory path the images should be saved to.")
	ipPtr := flag.StringP("ip-address", "a", "", "The IP MIS2 should bind to")
	portPtr := flag.IntP("port", "p", 29300, "The port MIS2 should bind to")

	flag.Parse()

	if !strings.HasSuffix(*dirPtr, "/") {
		*dirPtr = *dirPtr + "/"
	}

	ln, _ := net.Listen("tcp", *ipPtr+":"+strconv.Itoa(*portPtr))
	conn, _ := ln.Accept()

	/*reader := bufio.NewReader(conn)
	message, _ := reader.ReadString('\n')
	log.Println("Message Received:", string(message))

	conn.Write([]byte("thing"))*/

	image, err := png.Decode(conn)
	conn.Close()
	f, err := os.Create(*dirPtr + imageName() + ".png")
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, image)
	if err != nil {
		panic(err)
	}
	f.Close()
}

const allowedCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

var src = rand.NewSource(time.Now().UnixNano())

func imageName() string {
	b := make([]byte, 5)
	for i, cache, remain := 4, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(allowedCharacters) {
			b[i] = allowedCharacters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

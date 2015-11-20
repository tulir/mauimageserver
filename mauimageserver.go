package main

import (
	"bufio"
	"fmt"
	flag "github.com/ogier/pflag"
	"image/png"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var dirPtr = flag.StringP("directory", "d", "./%s", "The directory path the images should be saved to. %s is replaced by the file name.")
var addrPtr = flag.StringP("image-address", "i", "http://localhost/%s", "The address where the images are available online. %s is replaced by the file name.")
var pwdPtr = flag.StringP("password", "w", "maumau", "The MIS2 password")
var ipPtr = flag.StringP("ip-address", "a", "", "The IP MIS2 should bind to")
var portPtr = flag.IntP("port", "p", 29300, "The port MIS2 should bind to")

func main() {
	flag.Parse()

	if !strings.HasSuffix(*dirPtr, "/") {
		*dirPtr = *dirPtr + "/"
	}
	ln, _ := net.Listen("tcp", *ipPtr+":"+strconv.Itoa(*portPtr))

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn, *pwdPtr)
	}
}

func handleConnection(conn net.Conn, pwd string) {
	reader := bufio.NewReader(conn)
	message, _ := reader.ReadString('|')
	message = strings.TrimSpace(message)

	if message != pwd {
		log.Println(conn.RemoteAddr().String() + " failed authentication (" + message + ")")
		conn.Write([]byte("false"))
		conn.Close()
		return
	}
	conn.Write([]byte("true"))

	name := imageName() + ".png"
	image, err := png.Decode(conn)
	if err != nil {
		panic(err)
	}
	conn.Write([]byte(fmt.Sprintf(*addrPtr, name)))
	conn.Close()
	f, err := os.Create(fmt.Sprintf(*dirPtr, name))
	if err != nil {
		panic(err)
	}
	err = png.Encode(f, image)
	if err != nil {
		panic(err)
	}
	f.Close()
	log.Println(conn.RemoteAddr().String() + " successfully uploaded an image to " + name)
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

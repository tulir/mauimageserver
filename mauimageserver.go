package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/ogier/pflag"
	"maunium.net/go/mauimageserver/data"
	log "maunium.net/go/maulogger"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var debug = flag.BoolP("debug", "d", false, "Enable to print debug messages to stdout")
var confPath = flag.StringP("config", "c", "./config.json", "The path of the mauImageServer configuration file.")

var config *data.Configuration

func main() {
	flag.Parse()

	// Configure the logger
	log.PrintDebug = *debug
	log.Fileformat = func(date string, i int) string { return fmt.Sprintf("logs/%[1]s-%02[2]d.log", date, i) }
	// Initialize the logger
	log.Init()

	log.Infof("Initializing mauImageServer")
	loadConfig()
	loadDatabase()

	log.Infof("Registering handlers")
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/", get)
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.ListenAndServe(config.IP+":"+strconv.Itoa(config.Port), nil)
}

func loadConfig() {
	log.Infoln("Loading config...")
	var err error
	config, err = data.LoadConfig(*confPath)
	if err != nil {
		log.Fatalf("Failed to load config: %[1]s", err)
		os.Exit(1)
	}
	log.Debugln("Successfully loaded config.")
}

func loadDatabase() {
	log.Infoln("Loading database...")

	var err error
	err = data.LoadDatabase(config.SQL)
	if err != nil {
		log.Fatalf("Failed to load database: %[1]s", err)
		os.Exit(2)
	}

	log.Debugln("Successfully loaded database.")
}

func getIP(r *http.Request) string {
	if config.TrustHeaders {
		return r.Header.Get("X-Forwarded-For")
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func output(w http.ResponseWriter, response interface{}, status int) bool {
	// Marshal the response
	json, err := json.Marshal(response)
	// Check if there was an error marshaling the response.
	if err != nil {
		// Write internal server error status.
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	w.WriteHeader(status)
	w.Write(json)
	return true
}

/*func handleConnection(conn net.Conn, pwd string) {
	reader := bufio.NewReader(conn)
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)

	if message != pwd {
		log.Println(conn.RemoteAddr().String() + " failed authentication (" + message + ")")
		conn.Write([]byte("false\n"))
		conn.Close()
		return
	}

	conn.Write([]byte("true\n"))

	name := imageName() + ".png"
	image, err := png.Decode(conn)
	if err != nil {
		panic(err)
	}
	conn.Write([]byte(fmt.Sprintf(*addrPtr, name) + "\n"))
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
}*/

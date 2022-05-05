package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var counter int = 0
var waitGroup sync.WaitGroup

func get(w http.ResponseWriter, req *http.Request) {
	log.Printf("get received: %v", counter)
	fmt.Fprintf(w, "got: %d\n", counter)
}

func set(w http.ResponseWriter, req *http.Request) {
	log.Printf("set %v", req)
	val := req.URL.Query().Get("val")
	intval, err := strconv.Atoi(val)

	if err != nil {
		log.Printf("error converting invalid number string to int: %s", err)
		return
		//panic("unhandled error")
	}

	counter = intval
	log.Printf("set to: %v", counter)
}

func inc(_ http.ResponseWriter, _ *http.Request) {
	waitGroup.Add(1)
	defer waitGroup.Done()

	go doIncrement()

	waitGroup.Wait()
	log.Printf("incremented to: %v", counter)
}

func doIncrement() {
	counter = counter + 1
}

func main() {
	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/increment", inc)

	portnum := 8000
	if len(os.Args) > 1 {
		portnum, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Printf("error converting invalid port argument %s to int: %s", os.Args[1], err)
			log.Printf("switching back to port %d", portnum)
		}
	}

	log.Printf("Going to listen on port %d\n", portnum)
	log.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(portnum), nil))
}

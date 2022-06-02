package main

import (
	"fmt"
	"log"
	"net/http"

	//might contain some interesting/useful things (I also like that you can comment these package entries)
	//"net/http/httputil"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

//Golang's platform independent int doesn't work with the atomic AddUint64, LoadUint64
var counter uint64 = 0

//below is used to synchronize multiple/repeated goroutines called from this app
//(not helpful for locking the http increment requests, but useful elsewhere)
//var waitGroup sync.WaitGroup

//As an option to using mutex, will look into using the sync features of goroutines and channels next time.
//Implementing thread-safe locking because without it, the increment is out of sequence.
//Tested with two simultaneous ApacheBench sessions - the results were a much faster execution time since it seemed race condition handling slowed
//things down prior to locking the threads (maybe)
//Go has a mutex feature for locking threads - not surprising, given its built-in concurrency architecture:
var mutex sync.Mutex

func get(w http.ResponseWriter, req *http.Request) {
	log.Printf("get received: %v", counter)
	fmt.Fprintf(w, "got counter value: %d\n", counter)
}

func set(w http.ResponseWriter, req *http.Request) {
	//log.Printf("set %v", req) //??
	val := req.URL.Query().Get("val")
	intval, err := strconv.ParseUint(val, 10, 64)

	if err != nil {
		//maybe not the correct way to do this, but I like te gist of it better than panic()
		//I'd have to get used to using the rather basic (?) Go error handling - I'm so used to try/catch/finally logic flow and Exception bubbling
		//(what does Go's call stack look like?)
		//if the http request querystring value is invalid (NaN), nothing can be processed, so terminate the func
		//%s and error may very well be incorrect here (too much info - pare it down to basic error message)
		log.Printf("error converting invalid number string to int: %s", err)
		return
		//panic("unhandled error")
	}

	//respect mutex for setting variable as well:
	mutex.Lock()
	defer mutex.Unlock()

	//set variable using atomic to maintain consistent, atomic operations for all participants of this set/increment API:
	atomic.SwapUint64(&counter, intval)
	//counter = intval

	log.Printf("set to: %v", counter)
	//fixed: wasn't sending response to the caller:
	fmt.Fprintf(w, "set counter to: %d\n", counter)
}

func inc(_ http.ResponseWriter, req *http.Request) {
	//each http request is handled by a unique concurrent goroutine, so the fine-tuned concurrency in Go by design as at play here,
	//and also taking into account that
	//MSFT's Web API also handles http requests concurrently, but also has a notion of SessionState, with various locking and request timeout rules
	//nodejs http server "feels" like it's similar to Go's library - will have to review

	//we need a mutual exclusion mechanism in order to increment contiguously in Go's concurrent world
	//(out of sequence increment values could be caused by the way the ab utility calls the http server)
	//either way, ensuring atomic, locking increment is quite straight forward in the Go paradigm:
	mutex.Lock()
	defer mutex.Unlock()

	//waitGroup.Add(1) //again.leaving this in for future reference
	atomic.AddUint64(&counter, 1) // makes the increment atomic, but is still lockless
	//counter += 1 //can also work as long as the mutex is used (don't need LoadUint64)
	//waitGroup.Done()

	//identify 'who' is making the http call:
	userAgent := req.Header.Get("User-Agent")

	//using atomic increment AND mutex (various solutions for different use cases come to mind)
	log.Printf("incremented to: %v by caller: %s", atomic.LoadUint64(&counter), userAgent)
	//log.Printf("incremented to: %v", counter)
	//waitGroup.Wait()
}

func main() {
	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/increment", inc)

	portnum := 8000
	if len(os.Args) > 1 {
		portnum, err := strconv.Atoi(os.Args[1])
		if err != nil {
			//same sort of idea, if the port argument is NaN, can't use it
			log.Printf("error converting invalid port argument %s to int: %s", os.Args[1], err)
			log.Printf("continuing with default port %d", portnum)
		}
	}

	log.Printf("Going to listen on port %d\n", portnum)
	log.Fatal(http.ListenAndServe("localhost:"+strconv.Itoa(portnum), nil))
}

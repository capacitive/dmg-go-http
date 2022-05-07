# Golang and concurrency
After looking at this code the first time, a few things came to mind:

1. If the requirement is to have the counter incremented contiguously in a global space for all callers, then a mutex or stateful goroutines/channels is what's required to make that work ('everyone' sees the same increment, and the counter value is communicated in correct order).  Sychronizing the thread execution is what we want.
2. If the goal is for every caller/user to have their own counter value per session, then some other approaches are required, which are too large in scope to tackle with the time allotted to this task. Compartmentalizing the thread execution is what we want.

## The path I chose
So I went with the paradigm of the singleton blockchain as global ledger.  Although it's not a singleton instance per se, each actor/validator/smart contract app works with an exact immutable copy of that blockchain (an ordered, back-linked list of stored transactions), regardless of sharding or how many copies there are.  The counter value is now the same for everyone who participates in that session (destroyed once the http server is terminated). The 

When I first ran the curl requests, everything looked normal in terms of counter increment order.  When I ran the ab test, the incrementing got out of order quickly.  Adding in the object locking mechanism using Go's mutex fixed this, and was very easy to implement.  Go is concurrent out of the box!  So Go is procedural but not thread blocking.  On top of all that, http requests (REST or otherwise), inherently create race conditions when a worker process is created for each incoming request (what I suspect http.HandleFunc is doing, not unlike .Net's Web API, nginx and Apache http server). So...

`var mutex sync.Mutex`

Then in the inc() function, you do:

```
mutex.Lock()
defer mutex.Unlock()
```

_then_ you increment the counter, and the concurrent processes making http requests have to wait for the mutex to unlock the increment execution.

Why did I use atomic increment for the counter?
-----------------------------------------------
To demonstrate another approach to the problem of managing/sychronizing state.  This is accompanied by the WaitGroup mechanism, and may be useful for certain use cases.  The usual path is using stateful goroutines and channels, which I've also demonstrated. Also, they didn't fix the race condition problem, and I didn't have time to fully investigate.

A different style
-----------------
On to stateful goroutines and channels!  To quote, "(the) channel-based approach aligns with Go’s ideas of sharing memory by communicating and having each piece of data owned by exactly 1 goroutine". Sounds like a plan. So instead of a mutex I could have used channels (?).However, this approach seems like overkill to me for this specific task - the mutex is much simpler. And with this approach the increments were still out of sequence, unable to prevent the race conditions caused by the multiple http requesters.  With some of the ways I tried this, the numbers were almost perfect, but there were misses and duplicates in the log output.  Most of the time it was no better than simply counter += 1 updating the global counter variable (which always results in incorrect sequence). In light of mutex's ease of use, this was not worth pursuing.  Semaphores?  Now I'm getting carried away with options I don't know well enough in specific Go syntax.

Other minor things I fixed/added
------------------------
I changed the platform-dependent int to Uint in order for the increment to work with the atomic functions.  The conversion of the querystring to int (Atoi) returns an int, and the atomic functions don't appear to support it. Maybe there's more to it that I missed, but time contraints drove the need to find the most immediate fix.  The second reason is that with only an increment function, we don't need an integer type that supports negative numbers.  I also went with 64 bit to keep big numbers (hash puzzles) and word sizes (performance) in mind that support blockchain technologies.

I also changed some of the error handling to get it on its way to a production environment. There's much more to do in that space.

I added the User-Agent to the logging to show that Go http library has some useful things in it for developing robust web apps (the http utility library looks interesting as well).

I planned to add in the channels example as a module, but how to do that seemed a bit unclear with my limited knowledge.  And I didn't want to risk development environment issues on your end.  So I added the channels example as a non-Go file for reference.

The set() handler had some extraneous logging, and wasn't sending a response to the caller.

Note: I've split up this work over many small pockets of time over the past few days, at the risk of producing something disjointed.


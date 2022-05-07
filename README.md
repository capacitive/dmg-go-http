Golang and concurrency
----------------------
After looking at this code the first time, a few things come to mind:

1. If the requirement is to have the counter incremented contiguously in a global space for all callers, then a mutex or stateful goroutines/channels is what's required to make that work ('everyone' sees the same increment, and the counter value is communicated in correct order).
2. If the goal is for every caller/user to have their own counter value per session, then some other approaches are required, which are too large in scope to tackle with the time allotted to this task.

The path I chose
----------------
----------------

So I went with the paradigm of the singleton blockchain as global ledger.  Although it's not a singleton instance per se, each actor/validator/smart contract app works with an exact immutable copy of that blockchain (an ordered, back-linked list of stored transactions), regardless of sharding or how many copies there are.  The counter value is now the same for everyone who participates in that session (destroyed once the http server is terminated)

When I first ran the curl requests, everything looked normal in terms of counter increment order.  When I ran the ab test, the incrementing got out of order quickly.  Adding in the object locking mechanism using Go's mutex fixed this, and was very easy to implement.  Go is concurrent out of the box!  So Go is procedural but not thread blocking.  On top of all that, http requests (REST or otherwise), inherently create race conditions if not handled, the name of this problem wer'e talking about. So...

`var mutex sync.Mutex`

Then in the inc() function, you do:


```
mutex.Lock()
defer mutex.Unlock()
```

_then_ you increment the counter.

Why did I use atomic increment for the counter?
-------------------------------------------------------
To demonstrate another approach to the problem of managing state.  This is accompanied by the WaitGroup mechanism, and may be useful for certain use cases.  The usual path is using stateful goroutines and channels, which I've also demonstrated.

A different style
-----------------
On to stateful goroutines and channels!








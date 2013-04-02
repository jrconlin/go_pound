# Go Pound

A go based websocket pounder for SimplePush

---

I needed a way to stress test the SimplePush server to see how many
simultaneous sockets it could hold. Figured that it's a good enough
reason to muck with Go. Behold my crap-tastic implementation.

This requires a config.json file that looks like:

    {
    "target": "ws://host:port/ws",
    "clients": 14000,
    "sleep": "10s"
    }

:target; the machine to test
:clients; the number of clients to spawn (note, >13000 go routines can be... problematic.
:sleep; time to sleep before sending the fake "ping" packet.


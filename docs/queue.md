# queue

## resources

- [queue-patterns.go - thrwan01](https://github.com/thrawn01/queue-patterns.go)

```go
func runLoop() {
	for {
		select {
		// Collect all the requests into a local queue
		case req := <-m.requestCh:
			// create container of batched requests
		EMPTY:
			for {
				select {
				case req := <-m.requestCh: // append to container
				default: break EMPTY
				}
			}
			// do work
		case <-m.done: return
		}
	}
}
```

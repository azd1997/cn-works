package main

import "runtime"

func main() {
	//StartHTTPProxy("", "8080")
	//StartProxy("", "8081", false, true)
	StartConcurrentProxy(
		"", "8081", false, true,
		runtime.NumCPU())
}

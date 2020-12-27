package main

import "fmt"

var VersionString = "no version info"

func main() {

	// A global timeout for an http server may not be the best fit
	// for all scenarios. I worked on streaming APIs in the past and
	// the stream can be long lived (both audio/media and also documents like
	// a JSON stream). So a config like that must be used with care to
	// not cause very odd bugs (like streams being cut short automatically).
	fmt.Printf("loaner version: %q\n", VersionString)
}

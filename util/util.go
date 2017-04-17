package util

import "log"
import "os"

// HandleErr : Prints out error when it happens
func HandleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// HardHandleErr : Prints out error when it happens
// Also exit the whole program
func HardHandleErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

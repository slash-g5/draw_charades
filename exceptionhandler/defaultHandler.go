package exceptionhandler

import "log"

func HandleErr(e error) bool {
	if e != nil {
		log.Fatal(e)
		return true
	}
	log.Default()
	return false
}

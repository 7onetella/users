package main

import "log"

func LogError(err error) {
	if err != nil {
		log.Printf("err while getting URL: %v", err)
	}
}

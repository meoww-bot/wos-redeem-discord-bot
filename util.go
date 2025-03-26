package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func PrettyPrint(data any) {
	var p []byte
	var err error
	p, err = json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
	log.Printf("%s \n", p)
}

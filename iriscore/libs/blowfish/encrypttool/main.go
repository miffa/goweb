package main

import (
	xxx "dbms/lib/blowfish"
	"flag"
	"fmt"
)

var demo int
var message string
var ifen bool
var ifde bool

func init() {
	flag.IntVar(&demo, "demo", 20000, "msgs per  second")
	flag.StringVar(&message, "message", "", "message to publish")
	flag.BoolVar(&ifen, "encrypt", false, "compress the messages published")
	flag.BoolVar(&ifde, "decrypt", false, "compress the messages published")
}

var (
	HELLOWORLD = []byte("Welcome2ucar~")
	NIHAOUCAR  = []byte("devops2015")
)

func main() {
	flag.Parse()

	var err error
	var uguess string
	var style string

	switch {
	case ifen && ifde:
	case !ifen && !ifde:
		fmt.Printf("Can i help you?\n")
		return
	case message == "":
		fmt.Printf("no text to xxoo\n")
		return
	}

	switch {
	case ifen:
		uguess, err = xxx.Encrypt(message, HELLOWORLD, NIHAOUCAR)
		style = "encrypt"
	case ifde:
		uguess, err = xxx.Decrypt(message, HELLOWORLD, NIHAOUCAR)
		style = "decrypt"
	}

	if err != nil {
		fmt.Printf("%s err:%v\n", style, err)
		return
	}

	fmt.Printf("%s %s output |%s|\n", message, style, uguess)
}

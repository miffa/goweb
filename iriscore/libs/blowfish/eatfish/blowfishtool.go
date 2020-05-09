package main

import (
	"dbms/lib/blowfish"
	"fmt"
	"os"
)

const usage = `
        1) command to encrypt: "exe encrypt/e text saiferkey salt"
            -- result --
            Encrypting "test"
            BlowFish Encrypted String: AAAAAAAAAABYm+3y9AZZzQ==
        2) command to decrypt: "exe decrypt/d AAAAAAAAAABYm+3y9AZZzQ== saiferkey salt"
            -- result --
            Decrypting "AAAAAAAAAABYm+3y9AZZzQ=="
            BlowFish Decrypted String: test
`

func main() {
	// default string to encrypt
	var strToEncrDcr string
	var cf string
	var salt string
	var whatami string
	strArg := os.Args[1:]
	if len(strArg) < 3 {
		fmt.Println(usage)
		return

	}
	whatami = strArg[0]
	strToEncrDcr = strArg[1]
	if strToEncrDcr == "" {
		fmt.Println("text is empty")
		return
	}
	cf = strArg[2]
	if len(strArg) == 4 {
		salt = strArg[3]
	}

	switch whatami {
	case "e":
		fallthrough
	case "encrypt":
		fmt.Println("Encrypting \"" + strToEncrDcr + "\"")
		encr, err := blowfish.Encrypt(strToEncrDcr, []byte(cf), []byte(salt))
		if err != nil {
			fmt.Println("Encrypt err:", err)
			return
		}
		fmt.Println("BlowFish Encrypted String:[" + encr + "]")

	case "d":
		fallthrough
	case "decrypt":
		fmt.Println("Decrypting \"" + strToEncrDcr + "\"")
		encr, err := blowfish.Decrypt(strToEncrDcr, []byte(cf), []byte(salt))
		if err != nil {
			fmt.Println("Decrypt err:", err)
			return
		}
		fmt.Println("BlowFish Decrypted String:[" + encr + "]")
	default:
		fmt.Println("What are you doing (Encrypte or Decrypt)")
		fmt.Println(usage)
	}

}

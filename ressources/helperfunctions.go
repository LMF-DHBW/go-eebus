package ressources

import (
	"fmt"
	"os"
	"runtime/debug"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func CheckError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		debug.PrintStack()
		os.Exit(1)
	}
}

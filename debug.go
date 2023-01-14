// debug.go

package main

import (
	"fmt"
	"os"
	//"runtime"
	//"path/filepath"
	// "github.com/git719/aza"
	// "github.com/git719/utl"
)
var (
	// Some basic colors
	Red = "\033[1;31m" ; Gre = "\033[0;32m" ; Blu = "\033[1;34m"
	Yel = "\033[0;33m" ; Pur = "\033[1;35m" ; Cya = "\033[0;36m"
	Rst = "\033[0m"
)

// func init() {
// 	if runtime.GOOS == "windows" {
// 		Red = "" ; Gre = "" ; Blu = "" ; Yel = "" ; Pur = "" ; Cya = "" ; Rst = ""
// 	}
// }

func TestFunction() {
	fmt.Println(Red + "Hellow world!" + Rst)
	fmt.Println(Gre + "Hellow world!" + Rst)
	fmt.Println(Blu + "Hellow world!" + Rst)
	fmt.Println(Yel + "Hellow world!" + Rst)
	fmt.Println(Pur + "Hellow world!" + Rst)
	fmt.Println(Cya + "Hellow world!" + Rst)

	// var z aza.AzaBundle

// 	// Set up the bundle of variables
// 	z = SetupVariables(&z)
// 	fmt.Println(z.ConfDir)
// 	fmt.Println(z.CredsFile)
// 	fmt.Println(z.TokenFile)
	
// 	// Setup the API tokens
// 	z = aza.SetupApiTokens(&z)

// 	// Get subscription with a specific filter
// 	subs := GetSubscriptions("as01", z)
// 	fmt.Println(len(subs))
// 	utl.PrintJson(subs)
	os.Exit(0)
}

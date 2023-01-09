// debug.go

package main

// import (
// 	"fmt"
// 	"os"
// 	"github.com/git719/aza"
// 	"github.com/git719/utl"
// )

// func TestFunction() {
// 	var z aza.AzaBundle
// 	var oMap MapString

// 	// Set up the bundle of variables
// 	z, oMap = SetupVariables(&z, &oMap)
// 	fmt.Println(z.ConfDir)
// 	fmt.Println(z.CredsFile)
// 	fmt.Println(z.TokenFile)
// 	fmt.Println(oMap["d"])
// 	fmt.Println(oMap["m"])
// 	fmt.Println(oMap["sp"])
	
// 	// Setup the API tokens
// 	z = aza.SetupApiTokens(&z)

// 	// Get subscription with a specific filter
// 	subs := GetSubscriptions("as01", z)
// 	fmt.Println(len(subs))
// 	utl.PrintJson(subs)
// 	os.Exit(0)
// }

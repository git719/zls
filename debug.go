// debug.go

package main

import (
	"fmt"
	"os"
	//"path/filepath"
	// "github.com/git719/maz"
	"github.com/git719/utl"
)

func TestFunction() {
	fmt.Println(utl.Red + "Hellow world!" + utl.Rst)
	fmt.Println(utl.Gre + "Hellow world!" + utl.Rst)
	fmt.Println(utl.Blu + "Hellow world!" + utl.Rst)
	fmt.Println(utl.Yel + "Hellow world!" + utl.Rst)
	fmt.Println(utl.Pur + "Hellow world!" + utl.Rst)
	fmt.Println(utl.Cya + "Hellow world!" + utl.Rst)

	// var z maz.Bundle

	// 	// Set up the bundle of variables
	// 	z = SetupVariables(&z)
	// 	fmt.Println(z.ConfDir)
	// 	fmt.Println(z.CredsFile)
	// 	fmt.Println(z.TokenFile)

	// 	// Setup the API tokens
	// 	z = maz.SetupApiTokens(&z)

	// 	// Get subscription with a specific filter
	// 	subs := GetSubscriptions("as01", z)
	// 	fmt.Println(len(subs))
	// 	utl.PrintJson(subs)
	os.Exit(0)
}

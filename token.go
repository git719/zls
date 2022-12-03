// token.go

package main

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func GetToken(scopes []string) (token string, err error) {
	// Set up token cache storage file and accessor
	cache_file := filepath.Join(confdir, "accessTokens.json")
	cacheAccessor := &TokenCache{cache_file}

	if interactive == "true" {
		// Interactive login with 'public' app
        // See https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/dev/apps/public/public.go

		// Interactive login uses the 'Azure PowerShell' client_id
		psClientId := "1950a258-227b-4e31-a9cf-717495945fc2"  // Using a local variable for this
		// See https://stackoverflow.com/questions/30454771/how-does-azure-powershell-work-with-username-password-based-auth
		app, err := public.New(psClientId, public.WithAuthority(authority_url), public.WithCache(cacheAccessor))
		if err != nil {
			panic(err.Error())
		}

		// Select the account to use based on username variable 
		var targetAccount public.Account  // Type is defined in 'public' module 
		for _, i := range app.Accounts() {
			if strings.ToLower(i.PreferredUsername) == username {
				targetAccount = i
				break
			}
		}

		// Try getting cached token 1st
		result, err := app.AcquireTokenSilent(context.Background(), scopes, public.WithSilentAccount(targetAccount))
		if err != nil {
			// Else, get a new token
			result, err = app.AcquireTokenInteractive(context.Background(), scopes)
			// AcquireTokenInteractive acquires a security token from the authority using the default web browser to select the account.
			if err != nil {
				panic(err.Error())
			}
		}
		return result.AccessToken, nil	// Return only the AccessToken, which is of type string
	} else {
		// Client_id + secret login automated login with 'confidential' app
		// See See https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/dev/apps/confidential/confidential.go

		// Initializing the client credential
		cred, err := confidential.NewCredFromSecret(client_secret)
		if err != nil {
			print("Could not create a cred object from client_secret.\n")
		}

		// Automated login obviously uses the registered app client_id (App ID)
		app, err := confidential.New(client_id,	cred, confidential.WithAuthority(authority_url), confidential.WithAccessor(cacheAccessor))
		if err != nil {
			panic(err.Error())
		}

		// Try getting cached token 1st
		// targetAccount not required, as it appears to locate existing cached tokens without it
		result, err := app.AcquireTokenSilent(context.Background(), scopes)
		if err != nil {
			// Else, get a new token
			result, err = app.AcquireTokenByCredential(context.Background(), scopes)
			// AcquireTokenByCredential acquires a security token from the authority, using the client credentials grant.
			if err != nil {
				panic(err.Error())
			}
		}
		return result.AccessToken, nil  // Return only the AccessToken, which is of type string
	}
}

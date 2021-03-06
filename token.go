// token.go

package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

func GetToken(scopes []string) (token string, err error) {
	// Initializing the client credential
	cred, err := confidential.NewCredFromSecret(client_secret)
	if err != nil {
		fmt.Println("Could not create a cred from client_secret.")
	}

	// Set up token cache storage file and accessor
	f := filepath.Join(confdir, "accessTokens.json")
	cacheAccessor := &TokenCache{f}

	app, err := confidential.New(client_id,
		cred,
		confidential.WithAuthority(authority_url),
		confidential.WithAccessor(cacheAccessor))
	if err != nil {
		panic(err.Error())
	}

	// Try getting cached token 1st
	result, err := app.AcquireTokenSilent(context.Background(), scopes)
	if err != nil {
		// Else, get a new token
		result, err = app.AcquireTokenByCredential(context.Background(), scopes)
		if err != nil {
			panic(err.Error())
		}
	}

	token = result.AccessToken // Return only the AccessToken, which is of type string
	return token, nil
}

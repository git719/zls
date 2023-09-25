# pman
Simple Azure REST API caller. It uses the token the `zls` utility is able to acquire with arguments `-tmg` and `-taz`.

## Get Started
Copy the `pman` BASH script to somewhere in your path, maybe `/usr/local/bin/`

The script should work in any BASH environment that is able to run `curl` and the `zls` utility.

## Usage
```
$ pman
pman Azure REST API Caller v1.0.0
  Usage Examples:
    pman GET "https://graph.microsoft.com/v1.0/me"
    pman GET "https://management.azure.com/subscriptions?api-version=2022-12-01" [other curl options]
    pman GET "https://graph.microsoft.com/v1.0/applications/3eec32f4-6ca1-4d1d-9335-19518aa196c4" [other curl options]
```

```
$ pman GET "https://graph.microsoft.com/v1.0/me" | jq
{
  "@odata.context": "https://graph.microsoft.com/v1.0/$metadata#users/$entity",
  "businessPhones": [],
  "displayName": "First Last",
  "givenName": "First",
  "jobTitle": null,
  "mail": null,
  "mobilePhone": null,
  "officeLocation": null,
  "preferredLanguage": "en",
  "surname": "Last",
  "userPrincipalName": "First@mydomain.onmicrosoft.com",
  "id": "12c12716-354b-49e2-9a20-d2bc07d730df"
}
```

# zls
Azure RBAC and MS Graph listing utility.

This utility is more a proof-of-concept in using the [MSAL library for Go](https://github.com/AzureAD/microsoft-authentication-library-for-go), but can be used for casual querying of certain Azure ARM and MS Graph API objects.

This is the GoLang version. The Python version is called `azls` and sits within the https://github.com/git719/az repo. This version is a little snappier, and the single binary executable is easier to use without having to setup a Python environment. 

Why `zls`? Because three-letter names are easier to type.

## Requirements
You must register a client app in your tenant and grant it the required Read permissions for all API object types this utility tries to list. Please find other sources for how to do the app reg.

The utility sets up and uses a configuration directory at `$HOME/.<utility_name>` to retrieve and store the required credential parameters, and also to store local cache files. The `credentials.json` file must be formated as follows:
```
{
    "tenant_id" : "UUID",
    "client_id" : "UUID",
    "client_secret" : "SECRET"
}
```
If `credentials.json` file doesn't exist, an empty skeleton one will be created that can be filled out accordingly.

## Getting started
To compile, you obviously need to have GoLang installed and properly setup in your system, with `$GOPATH` set up correctly (typically at `$HOME/go`). Also setup `$GOPATH/bin/` in your `$PATH`, since that's where the executable binary will be placed.

From a `bash` shell type `./build` to build the binary. 

This utility has been successfully tested in macOS, Linux, and Windows. To build from a regular Windows Command Prompt, just run the corresponding line in the `build` file (`go build ...`).

## Known Issues
Unknown.

## Usage
```
zls Azure RBAC and MS Graph listing utility v160
    -Xc                List total number of X objects in tenant
    -Xj                List all X objects in JSON format
    -X                 List all X objects tersely (UUID and most essential attributes)
    -X "string"        List all X objects whose name has "string" in it
    -Xj UUID|"string"  List specific X or matching objects in JSON format
    -X UUID            List specific X object in YAML-like human-readable format
    -X <specfile>      Compare X object specfile to what's in Azure
    -Xx                Delete cached X object local file

    Where 'X' can be any of these object types:
      'd'  = Role Definitions   'a'  = Role Assignments   's'  = Azure Subscriptions
      'm'  = Management Groups  'u'  = Azure AD Users     'g'  = Groups
      'sp' = Service Principals 'ap' = Applications

    -ar                List all role assignments with resolved names
    -mt                List Management Group and subcriptions tree
    -pags              List all Azure AD Priviledge Access Groups
    -tx                Delete cached accessTokens file
    -v                 Print this usage page
```

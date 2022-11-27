# zls
Azure RBAC and MS Graph listing utility.

This utility is more of a _proof-of-concept_ in how to use [MSAL library for Go](https://github.com/AzureAD/microsoft-authentication-library-for-go) than anything else. However, some may find it useful to casually query certain Azure ARM and MS Graph API objects.

This is the GoLang version. The Python version is called `azls` and sits within the https://github.com/git719/az repo. This version is a little snappier, and the single binary executable is easier to use without having to setup a Python environment. 

Why `zls`? Because three-letter names are easier to type.

## Requirements
You must register a client app in your tenant and grant it the required Read permissions for all API object types this utility tries to list. Please find other sources for how to do the app reg.

The utility sets up and uses a configuration directory at `$HOME/.<utility_name>` to retrieve and store the required credential parameters, and also to store local cache files. The `credentials.yaml` file must be formatted as follows:
```
tenant_id:     UUID-FOR-YOUR-TENANT
client_id:     UUID-FOR-YOUR-READER-REGISTERED-APP
client_secret: SECRET
```
You can also feed above 3 parameters to the utility using `-cr` option and it will create the `credentials.yaml` file. The interactive `-cri` login option is still being developed and currently NOT WORKING.

## Getting started
To compile, you obviously need to have GoLang installed and properly setup in your system, with `$GOPATH` set up correctly (typically at `$HOME/go`). Also setup `$GOPATH/bin/` in your `$PATH`, since that's where the executable binary will be placed.

From a `bash` shell type `./build` to build the binary. 

This utility has been successfully tested in macOS, Linux, and Windows. To build from a regular Windows Command Prompt, just run the corresponding line in the `build` file (`go build ...`).

## Object Types
This utility currently supports reading and reporting on the following Azure resource and Azure AD object types:

1. RBAC Role Definitions

2. RBAC Role Assignments

3. Azure Subscriptions

4. Azure Management Groups

5. Azure AD Users

6. Azure AD Groups

7. Service Principals

8. Applications

9. Azure AD Roles that have been **activated**

10. Azure AD Roles standard definitions

Additionally, it can also compare RBAC role definitions and assignments in a JSON or YAML _specification file_ to what that object currently looks like in the Azure tenant.

## Known Issues
Unknown.

## Usage
```
zls Azure RBAC and MS Graph listing utility v171
    -Xj                List all X objects in JSON format
    -X                 List all X objects tersely (UUID and most essential attributes)
    -X "string"        List all X objects whose name has "string" in it
    -Xj UUID|"string"  List specific X or matching objects in JSON format
    -X UUID            List specific X object in YAML-like human-readable format
    -X <specfile>      Compare X object specification file to what's in Azure
    -Xx                Delete X object cache local file

    Where 'X' can be any of these object types:
      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions
      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups
      sp = Service Principals      ap = Applications            ra = Azure AD Roles Active
      rd = Azure AD Roles Defs

    -ar                              List all RBAC role assignments with resolved names
    -mt                              List Management Group and subscriptions tree
    -pags                            List all Azure AD Privileged Access Groups
    -cr                              Dump values in credentials file
    -cr  TENANT_ID CLIENT_ID SECRET  Set up secret login
    -cri TENANT_ID CLIENT_ID         Set up interactive login (NOT WORKING)
    -st                              List local cache count and Azure count of all objects
    -tx                              Delete accessTokens cache file
    -xx                              Delete ALL cache local file
    -v                               Print this usage page
```

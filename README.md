# zls
Azure RBAC and MS Graph listing utility.

Test comment.

This utility is more of a _proof-of-concept_ in how to use [MSAL library for Go](https://github.com/AzureAD/microsoft-authentication-library-for-go) than anything else. However, some may find it useful to casually query certain Azure ARM and MS Graph API objects.

This is the GoLang version. The Python version is called `azls` and sits within the <https://github.com/git719/az> repo. This version is more much quicker, and having a single binary executable is easier to use without having to setup any Python environments.

Why `zls`? Because three-letter names are easier to type.

## Reader Access Requirements
1. You can register a dedicated Azure application in your tenant and grant it the required Read permissions for each API object type the utility reports on. Please find other sources for how to do the app reg. The utility sets up and uses a configuration directory at `$HOME/.<utility_name>` to retrieve and store the required credential parameters (`credentials.yaml`, but also to store other local cache files. To use this automated logon option, update the credentials file with the `-cr TENANT_ID CLIENT_ID SECRET` arguments once you have registered your special app for this.

2. Or you can do an MSAL interactive browser popup logon into your tenant, assuming your logon has the required Reader privileges. To use this option, update the credentials file by using `-cri TENANT_ID USERNAME` option.

## Getting started
To compile, you obviously need to have GoLang installed and properly setup in your system, with `$GOPATH` set up correctly (typically at `$HOME/go`). Also setup `$GOPATH/bin/` in your `$PATH`, since that's where the executable binary will be placed.

From a `bash` shell type `./build` to build the binary. 

This utility has been successfully tested on macOS, Ubuntu Linux, and Windows. To build from a regular Windows Command Prompt, just run the corresponding line in the `build` file (`go build ...`).

## Reported Object Types
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
> "Inside every large program is a small program struggling to get out." - Tony Hoare

## Utility Philosophy
Please note that the **primary goal** of this utility is to serve as a study aid for understanding the MSAL library as coded in the Go language. In addition, it can serve as a very useful little utility for permforming quick CLI checks of objects in your Azure cloud. The primary coding goal is to keep things very clear and simple to understand and maintain. 

## Usage
```
zls Azure RBAC and MS Graph listing utility v184
    -vs SPECFILE         Compare YAML or JSON specfile to what's in Azure (only for d and a objects)
    -X[j]                List all X objects tersely, with JSON output option
    -X[j] UUID|"string"  Show/list X object(s) matching on UUID or "string" attribute, JSON option
    -Xx                  Delete X object local file cache

    Where 'X' can be any of these object types:
      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions
      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups
      sp = Service Principals      ap = Applications            ra = Azure AD Roles Active
      rd = Azure AD Roles Defs

    -xx                               Delete ALL cache local files
    -ar                               List all RBAC role assignments with resolved names
    -mt                               List Management Group and subscriptions tree
    -pags                             List all Azure AD Privileged Access Groups
    -st                               List local cache count and Azure count of all objects

    -z                                Dump variables in running program
    -cr                               Dump values in credentials file
    -cr  TENANT_ID CLIENT_ID SECRET   Set up MSAL automated client_id + secret login
    -cri TENANT_ID USERNAME           Set up MSAL interactive browser popup login
    -tx                               Delete MSAL accessTokens cache file
    -v                                Print this usage page
```

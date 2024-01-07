# zls
`zls` is a [CLI](https://en.wikipedia.org/wiki/Command-line_interface) utility for **listing** Azure objects. It is similar to `az`, the official [Azure CLI tool](https://learn.microsoft.com/en-us/cli/azure/), but it is much **faster** because it is written in [Go](https://go.dev/) and compiled into a binary executable, and it also only focuses on a smaller set of Azure object types. It is a little _Swiss Army knife_ that can very **quickly** do the following:

- List the following [Azure Resources Services](https://que.tips/azure/#azure-resource-services) objects in your tenant:
  - RBAC Role Definitions
  - RBAC Role Assignments
  - Azure Subscriptions
  - Azure Management Groups
- List the following [Azure Security Services](https://que.tips/azure/#azure-security-services) objects:
  - Azure AD Users
  - Azure AD Groups
  - Service Principals
  - Applications
  - Azure AD Roles that have been **activated**
  - Azure AD Roles standard definitions
- Compare RBAC role definitions and assignments that are defined in a JSON or YAML __specification file__ to what that object currently looks like in the Azure tenant.
- Dump the current Resources or Security JWT token being used (which can be used as a [simple Azure REST API caller](https://github.com/git719/zls/tree/main/pman) for testing purposes) 
- Perform other related listing functions.

## Quick Example
A quick example is listing the Azure Built-in RBAC "Billing Reader" role:

```
$ zls -d "Billing Reader"
id: fa23ad8b-c56e-40d8-ac0c-ce449e1d2c64
properties:
  roleName: Billing Reader
  description: Allows read access to billing data
  assignableScopes:
    - /
  permissions:
    - actions:
        - Microsoft.Authorization/*/read
        - Microsoft.Billing/*/read
        - Microsoft.Commerce/*/read
        - Microsoft.Consumption/*/read
        - Microsoft.Management/managementGroups/read
        - Microsoft.CostManagement/*/read
        - Microsoft.Support/*
      notActions:
      dataActions:
      notDataActions:
```

- Another way of listing the same role is to call it by its UUID: `zls -d fa23ad8b-c56e-40d8-ac0c-ce449e1d2c64`
- The YAML listing format is more human-friendly and easier to read, and only displays the attributes that are most relevant to Azure systems engineers
- But if you wish to display it in JSON format you would simply use: `zls -dj fa23ad8b-c56e-40d8-ac0c-ce449e1d2c64`
- One advantage of the JSON format is that it displays every single attribute in the Azure object

## Introduction
The utility was primarily developed as a __proof-of-concept__ to:

- Learn to develop Azure utilities in the Go language
- Develop a small framework library for acquiring Azure [JWT](https://jwt.io/) token using the [MSAL library for Go](https://github.com/AzureAD/microsoft-authentication-library-for-go)
- Get a token for either an Azure user or a Service Principal (SP)
- Get the token to access the tenant's **Resources** Services API via endpoint <https://management.azure.com> ([REST API](https://learn.microsoft.com/en-us/rest/api/azure/))
- Get the token to access the tenant's **Security** Services API via endpoint <https://graph.microsoft.com> ([MS Graph](https://learn.microsoft.com/en-us/graph/overview))
- Do quick and dirty searches of any of the above mentioned object types in the azure tenant
- Develop small CLI utilities that call and use other Go library packages

This program was originally written in Python, a version called `azls`, which still sits within the <https://github.com/git719/az> repo. However, that effort was essentially abandoned because this Go version is **much quicker**, and having a simple binary executable is **much easier** than having to set up Python execution environments.

## Getting Started
To compile `zls`, first make sure you have installed and set up the Go language on your system. You can do that by following [these instructions here](https://que.tips/golang/#install-go-on-macos) or by following other similar recommendations found across the web.

- Also ensure that `$GOPATH/bin/` in your `$PATH`, since that's where the executable binary will be placed.
- Open a `bash` shell, clone this repo, then switch to the `zls` working directory
- Type `./build` to build the binary executable
- To build from a regular Windows Command Prompt, just run the corresponding line in the `build` file (`go build ...`)
- If there are no errors, you should now be able to type `zls` and see the usage screen for this utility.

This utility has been successfully tested on macOS, Ubuntu Linux, and Windows. In Windows it works from a regular CMD.EXE, or PowerShell prompts, as well as from a GitBASH prompt.

Below other sections in this README explain how to set up access and use the utility in your own Azure tenant. 

## Access Requirements
First and foremost you need to know the special **Tenant ID** for your tenant. This is a UUID that uniquely identifies your Microsoft Azure tenant.

Then, you need a User ID or a Service Principal with the appropriate access rights. Either one will need the necessary _Reader_ role access to read resource objects, and _Global Reader_ role access to read security objects. The higher the scope of these access assignments, the more you will be able to see with the utility. 

When you run `zls` without any arguments you will see the **usage** screen listed at the bottom of this README. As you can probably surmise, the `-cri` and `-cr` arguments will allow you to set up these 2 optional ways to connect to your tenant. The `-cri` argument mean set up 'credential interactive' for a User ID, also knows as a [User Principal Name (UPN)](https://learn.microsoft.com/en-us/entra/identity/hybrid/connect/plan-connect-userprincipalname), and the `-cr` argument means set up 'credential' for a Service Principal or SP.

## User ID Logon
For example, if your Tenant ID was **c44154ad-6b37-4972-8067-0ef1068079b2**, and your User ID was __bob@contoso.com__, you would type:

```
$ zls -cri c44154ad-6b37-4972-8067-0ef1068079b2 bob@contoso.com
Updated /Users/myuser/.maz/credentials.yaml file
```
`zls` responds that the special `credentials.yaml` file has been updated accordingly.

To view, dump all configured logon values type the following:

```
$ zls -z
config_dir: /Users/myuser/.maz  # Config and cache directory
os_environment_variables:
  # 1. Environment Variable login values override values in credentials_config_file
  # 2. MAZ_USERNAME+MAZ_INTERACTIVE login have priority over MAZ_CLIENT_ID+MAZ_CLIENT_SECRET login
  # 3. To use MAZ_CLIENT_ID+MAZ_CLIENT_SECRET login ensure MAZ_USERNAME & MAZ_INTERACTIVE are unset
  MAZ_TENANT_ID:
  MAZ_USERNAME:
  MAZ_INTERACTIVE:
  MAZ_CLIENT_ID:
  MAZ_CLIENT_SECRET:
credentials_config_file:
  file_path: /Users/myuser/.maz/credentials.yaml
  tenant_id: c44154ad-6b37-4972-8067-0ef1068079b2
  username: bob@contoso.com
  interactive: true
```

Above tells you that the utility has been configured to use Bob's UPN for access via the special credentials file. Note that above is only a configuration setup, it actually hasn't logged Bob onto the tenant yet. To logon as Bob you have have to run any command, and the logon will happen automatically, in this case it will be an interactively browser popup logon.

Note also, that instead of setting up Bob's login via the `-cri` argument, you could have setup the special 3 operating system environment variables to achieve the same. Had you done that, running `zls -z` would have displayed below instead:

```
$ zls -z
config_dir: /Users/myuser/.maz  # Config and cache directory
os_environment_variables:
  # 1. Environment Variable login values override values in credentials_config_file
  # 2. MAZ_USERNAME+MAZ_INTERACTIVE login have priority over MAZ_CLIENT_ID+MAZ_CLIENT_SECRET login
  # 3. To use MAZ_CLIENT_ID+MAZ_CLIENT_SECRET login ensure MAZ_USERNAME & MAZ_INTERACTIVE are unset
  MAZ_TENANT_ID: c44154ad-6b37-4972-8067-0ef1068079b2
  MAZ_USERNAME: bob@contoso.com
  MAZ_INTERACTIVE: true
  MAZ_CLIENT_ID:
  MAZ_CLIENT_SECRET:
credentials_config_file:
  file_path: /Users/myuser/.maz/credentials.yaml
  tenant_id: 
  username: 
  interactive:
```

## SP Logon
To use an SP logon it means you first have to set up a dedicated App Registrations, grant it the same Reader resource and Global Reader security access roles mentioned above. Please reference other sources on the Web for how to do an Azure App Registration.

Once above is setup, you then follow the same logic as for User ID logon above, but using `-cr` instead; or use the other environment variables (MAZ_CLIENT_ID andMAZ_CLIENT_SECRET ). 

The utility ensures that the permissions for configuration directory where the `credentials.yaml` file is only accessible by the owning user. However, storing a secrets in a clear-text file is a very poor security practice and should __never__ be use other than for quick tests, etc. The environment variable options was developed pricisely for this SP logon pattern, where the `zls` utility could be setup to run from say a [Docker container](https://en.wikipedia.org/wiki/Docker_(software)) and the secret injected as an environment variable, and that would be a much more secure way to run the utility.

These login methods and the environment variables are described in more length in the [maz](https://github.com/git719/maz) package README.

## To-Do and Known Issues
The program is stable enough to be relied on as a reading/listing utility, but there are a number of little niggly things that could be improved. Will put a list together at some point.

At any rate, no matter how stable any code is, it is always worth remembering computer scientist [Tony Hoare](https://en.wikipedia.org/wiki/Tony_Hoare)'s famous quote:
> "Inside every large program is a small program struggling to get out."

## Coding Philosophy
As mention in the *Introduction* above, the primary goal of this utility is to serve as a study aid for coding Azure utilities in Go, as well as to serve as a quick, _Swiss Army knife* utility to list tenant objects.

If you look through the code you will note that it is very straightforward. Keeping the code clear, and simple to understand and maintain is another coding goal.

Note that the bulk of the code is actually in the [maz](https://github.com/git719/maz) library, and other packages.

## Usage
```
zls Azure Resource RBAC and MS Graph READER v2.3.4
    READER FUNCTIONS
    UUID                              Show object for given UUID
    -vs Specfile                      Compare YAML or JSON specfile to what's in Azure (only for d and a objects)
    -X[j] [Specifier]                 List all X objects tersely, with option for JSON output and/or match on Specifier
    -Xx                               Delete X object local file cache

      Where 'X' can be any of these object types:
      d  = RBAC Role Definitions   a  = RBAC Role Assignments   s  = Azure Subscriptions
      m  = Management Groups       u  = Azure AD Users          g  = Azure AD Groups
      sp = Service Principals      ap = Applications            ad = Azure AD Roles

    -xx                               Delete ALL cache local files
    -ar                               List all RBAC role assignments with resolved names
    -mt                               List Management Group and subscriptions tree
    -pags                             List all Azure AD Privileged Access Groups
    -st                               List local cache count and Azure count of all objects

    -z                                Dump configured login values
    -zr                               Dump runtime variables
    -cr  TenantId ClientId Secret     Set up MSAL automated ClientId + Secret login
    -cri TenantId Username            Set up MSAL interactive browser popup login
    -tx                               Delete MSAL accessTokens cache file
    -tmg                              Dump current token string for MS Graph API
    -taz                              Dump current token string for Azure Resource API
    -tc "TokenString"                 Dump token claims
    -v                                Print this usage page
```

Instead of documenting individual examples of all of the above switches, it is best for you to play around with the utility to see the different listing functionality that it offers.

## Feedback
This utility along with the required libraries are obviously very useful to me, which is why I wrote them. However, I don't know if anyone else feels the same, which is why I have not yet thought about formalizing a proper feedback process.

The licensing is fairly open, so if you find it useful, feel free to clone and use on your own, with proper attributions. However, if you do see anything that could help improve any of this please let me know.

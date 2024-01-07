# zls
`zls` is a command-line (CLI) utility for **reading** Azure resource and security services.

It is similar to `az`, the official [Azure CLI tool](https://learn.microsoft.com/en-us/cli/azure/), but it is much faster because it is written in [Go](https://go.dev/) and compiled into an executable binary, and it also focuses on a smaller set of Azure services.

It can **quickly** display any of the following objects within an organization's Azure tenant:

- [Azure Resources Services](https://que.tips/azure/#azure-resource-services) objects:
  - RBAC Role Definitions
  - RBAC Role Assignments
  - Azure Subscriptions
  - Azure Management Groups
- [Azure Security Services](https://que.tips/azure/#azure-security-services) objects:
  - Azure AD Users
  - Azure AD Groups
  - Service Principals
  - Applications
  - Azure AD Roles that have been **activated**
  - Azure AD Roles standard definitions
- Additionally, it can also compare RBAC role definitions and assignments that are in a JSON or YAML __specification file__ to what that object currently looks like in the Azure tenant.

## Introduction
The utility was primarily developed as a __proof-of-concept__ to do the following using Go:

- Acquire an Azure [JWT](https://jwt.io/) token using the [MSAL library for Go](https://github.com/AzureAD/microsoft-authentication-library-for-go)
- Get the token using either an Azure user or a Service Principal (SP)
- Get the token to access the tenant's **Resources** Services via <https://management.azure.com>
- Get the token to access the tenant's **Security** Services via <https://graph.microsoft.com>
- Do quick and dirty searches of any of the above mentioned object types in the azure tenant

It was originally written in Python, a version called `azls`, which still sits within the <https://github.com/git719/az> repo. However, that effort has essentially been abandoned because this Go version is **much quicker**, and having a single binary executable is **much easier** than having to set up Python execution environments.

## Getting Started
To compile `zls`, first make sure you have installed and setup Go on your system. You can do that by following [these instructions here](https://que.tips/golang/#install-go-on-macos) or by following other similar recommendations found across the web.

- Also ensure that `$GOPATH/bin/` in your `$PATH`, since that's where the executable binary will be placed.
- From a `bash` shell 
- Clone this repo, and switch to the working directory
- Type `./build` to build the binary
- To build from a regular Windows Command Prompt, just run the corresponding line in the `build` file (`go build ...`)

This utility has been successfully tested on macOS, Ubuntu Linux, and Windows. In Windows it works from a regular CMD.EXE, or PowerShell prompts, as well as from a GitBASH prompt.

## Access Requirements
First and foremost you need to know the special **Tenant ID** for your tenant. This is a UUID that uniquely identifies your Microsoft Azure tenant.

Then, you need a User ID or a Service Principal with the appropriate access rights. Either one will need the necessary _Reader_ role access to read resource objects, and _Global Reader_ role access to read security objects. The higher the scope of these access assignments, the more you will be able to see with the utility. 

When you run `zls` without any arguments you will see the **usage** screen listed at the bottom of this README. As you can probably surmise, the `-cri` and `-cr` arguments will allow you to set up these 2 optional ways to connect. The arguments mean, set up 'credential interactive', for a User ID, or 'credential' for an SP.

## User ID Logon
For example, to logon to a bogus tenant with a Tenant ID of **c44154ad-6b37-4972-8067-0ef1068079b2**, and using bogus User ID __bob@contoso.com__, you would type:

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

Above tells you that the utility has been configured to use Bob's ID for access via the special credentials file. Note that above is only a configuration setup, it actually hasn't logged Bob onto the tenant yet. To logon as Bob you have have to run any command, and the logon will happen automatically, in this case it will be an interactively browser popup logon.

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

Once above is setup, you then follow the same logic as for User ID logon above, but using `-cr` instead; or use the other environment variables. 

The utility ensures that the permissions for configuration directory where the `credentials.yaml` file is kept is only accessible by the owning user. However, storing a secrets in clear-text is a very poor security practice and should not use other than for quick testing and so on. The environment variable options was developed pricisely for this SP logon logon pattern, when the `zls` utility can run from say a container, and the secret injected as an environment variable, and this would be much more secure.

These login methods and the environment variables are described in more length in the [maz](https://github.com/git719/maz) package.

## Known Issues
> "Inside every large program is a small program struggling to get out." - Tony Hoare

## Utility Philosophy
The primary goal of this utility is to serve as a study aid for understanding the MSAL library as coded in the Go language. In addition, it can serve as a very useful little utility for permforming quick CLI checks of objects in your Azure cloud. The coding philosophy is to keep things clear and simple to understand and maintain.

The bulk of the code is in the [maz](https://github.com/git719/maz) package.

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

## Examples
Instead of specific examples it is best to play around with the utility to see different ways of searching and printing different objects.

## Feedback and Contribution
I think this and other related projects are obviously very useful, which is why I wrote them. However, I don't know if anyone else feels the same, which is why I have not thought about formalizing a proper feedback and contribution process.

The licensing is fairly open, so if you find it useful, feel free to clone and use on your own, with proper attributions. But if you do see anything that could help improve this please let me know.

# aztoken.ps1

# Import required modules
Add-Type -AssemblyName System.Web

# Define colors
$BLUE = [System.ConsoleColor]::DarkBlue
$GREEN = [System.ConsoleColor]::DarkGreen
$RED = [System.ConsoleColor]::DarkRed
$RESET = [System.ConsoleColor]::White

# Initialize a global token cache
$cache = New-Object Microsoft.Identity.Client.TokenCache

# Define function to print with flushing
Function Print-Flush {
    param(
        [string]$message
    )
    Write-Host $message
    [System.Console]::Out.Flush()
}

# Define function to calculate expiry date
Function Expiry-Date {
    param(
        [int]$expiresInSeconds
    )
    $current_time = Get-Date
    if (-not $expiresInSeconds) {
        $expiresInSeconds = 0
    }
    $expiryDateTemp = $current_time.AddSeconds($expiresInSeconds)
    return $expiryDateTemp.ToString('yyyy-MM-dd HH:mm:ss')
}

# Define function to get token by credentials
Function Get-Token-By-Credentials {
    param(
        [string[]]$scopes,
        [string]$clientId,
        [string]$clientSecret,
        [string]$authorityUrl
    )
    # Define the client application using MSAL
    $cca = [Microsoft.Identity.Client.ConfidentialClientApplication]::new(
        $clientId,
        $authorityUrl,
        $clientSecret,
        $cache  # Use the global cache
    )

    # Acquire a token using client credentials
    $tokenRequest = @{
        'scopes' = $scopes
    }

    try {
        $result = $cca.AcquireTokenForClient($tokenRequest['scopes'])
        return $result
    } catch {
        throw "Error acquiring token: $result"
    }
}

# Define the main function
Function Main {
    # Define the scopes and other parameters
    $scopes = @('https://ossrdbms-aad.database.windows.net/.default')
    $clientId = $env:MAZ_CLIENT_ID
    $clientSecret = $env:MAZ_CLIENT_SECRET
    $tenantId = $env:MAZ_TENANT_ID
    $authorityUrl = "https://login.microsoftonline.com/$tenantId"

    while ($true) {
        $token = Get-Token-By-Credentials -scopes $scopes -clientId $clientId -clientSecret $clientSecret -authorityUrl $authorityUrl
        if (-not $token['access_token']) {
            Print-Flush "$RED Failed to obtain token: $token $RESET"
        }

        $access_token = $token['access_token']
        $expires_in_secs = $token['expires_in']
        $expires_in = Expiry-Date -expiresInSeconds $expires_in_secs

        Print-Flush ""
        Print-Flush "$BLUE TOKEN DETAILS$($RESET):"
        Print-Flush "$BLUE  client_id$($RESET) : $($GREEN)$($client_id)$($RESET)"
        Print-Flush "$BLUE  Authority$($RESET) : $($GREEN)$($authorityUrl)$($RESET)"
        Print-Flush "$BLUE  Scopes$($RESET)    : $($GREEN)$($scopes)$($RESET)"
        Print-Flush "$BLUE  Token$($RESET)     : $($GREEN)$($access_token)$($RESET)"
        Print-Flush "$BLUE  Expires On$($RESET): $($GREEN)$($expires_in)$($RESET) ($expires_in_secs seconds)"

        # Wait for 5 seconds (5,000 milliseconds) before making the next call
        Start-Sleep -Milliseconds 5000
    }
}

# Call the main function
Main

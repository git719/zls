// aztoken.js
// Node JS version

const msal = require('@azure/msal-node');
const { format } = require('date-fns');
const BLUE = '\x1b[1;34m'; // Blue color
const GREEN = '\x1b[32m';
const RED = '\x1b[1;31m';
const RESET = '\x1b[0m'; // Reset to default color
const dateFormatter = new Intl.DateTimeFormat('en-US', {
    timeZone: 'America/New_York',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
});
let cca;  // Declaring here to leverage MSAL's default in-memory cache


async function getTokenByCredentials(scopes, clientId, clientSecret, authorityUrl) {
    if (!cca) {
        cca = new msal.ConfidentialClientApplication({
        auth: {
            clientId: clientId,
            authority: authorityUrl,
            clientSecret: clientSecret
        }
        });
    }

    const tokenRequest = {
        scopes: scopes
    };

    try {
        const response = await cca.acquireTokenByClientCredential(tokenRequest);
        return response;
    } catch (error) {
        console.error('Error acquiring token:', error.message);
        throw error;
    }
}


async function main() {
    //const scopes = ['https://graph.microsoft.com/.default'];
    //const scopes = ['https://management.azure.com/.default'];
    const tokenScopes = ['https://ossrdbms-aad.database.windows.net/.default'];
    const clientId = process.env.MAZ_CLIENT_ID;
    const clientSecret = process.env.MAZ_CLIENT_SECRET;
    const authorityUrl = `https://login.microsoftonline.com/${process.env.MAZ_TENANT_ID}`;

    while (true) {
        try {
            const token = await getTokenByCredentials(tokenScopes, clientId, clientSecret, authorityUrl);

            // // Optional: Print entire token
            // console.log(`\n${token}\n`);
            
            const { authority, scopes, accessToken, fromCache, expiresOn, extExpiresOn } = token;
            const formattedExpiresOn = dateFormatter.format(new Date(expiresOn));
            const formattedExtExpiresOn = dateFormatter.format(new Date(extExpiresOn));

            console.log(`\n${BLUE}TOKEN DETAILS:${RESET}`);
            console.log(`${BLUE}  Authority:${RESET} ${GREEN}${authority}${RESET}`);
            console.log(`${BLUE}  Scopes:${RESET} ${GREEN}${scopes}${RESET}`);
            console.log(`${BLUE}  Access Token:${RESET} ${GREEN}${accessToken}${RESET}`);
            console.log(`${BLUE}  From Cache:${RESET} ${GREEN}${fromCache}${RESET}`);
            console.log(`${BLUE}  Expires On:${RESET} ${GREEN}${formattedExpiresOn}${RESET}`);
            console.log(`${BLUE}  Extended Expires On:${RESET} ${GREEN}${formattedExtExpiresOn}${RESET}`);              
        } catch (error) {
            console.error(`${RED}Failed to obtain token: ${error}${RESET}`);
        }

        // Wait for 1 minute (60,000 milliseconds) before making the next call
        await new Promise(resolve => setTimeout(resolve, 60000));
    }
}

main().catch(error => {
    console.error('An error occurred in the main function:', error);
});

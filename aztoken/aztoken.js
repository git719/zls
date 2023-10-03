// aztoken.js

// Quick exit on CTLR-C
process.once('SIGTERM', () => process.exit(0)).once('SIGINT', () => process.exit(0));

const msal = require('@azure/msal-node');
const { format } = require('date-fns');

const BLUE = '\x1b[1;34m'; // Blue color
const GREEN = '\x1b[32m';
const RED = '\x1b[31m';  // Removed 1;
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
        console.error(`${RED}Error acquiring token: ${error}${RESET}`);
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
    console.log(`\n${BLUE}clientId${RESET}: ${GREEN}${clientId}${RESET}`);
    console.log(`${BLUE}clientSecret${RESET}: ${GREEN}secret${RESET}`);
    console.log(`${BLUE}authorityUrl${RESET}: ${GREEN}${authorityUrl}${RESET}`);

    while (true) {
        const token = await getTokenByCredentials(tokenScopes, clientId, clientSecret, authorityUrl);
        if (!token || token === '') {
            console.error(`${RED}Token is empty or null.${error}${RESET}`);
        }

        // console.log(`\n${token}\n`);  // OPTION: Print entire token structure
        
        const { authority, scopes, accessToken, fromCache, expiresOn, extExpiresOn } = token;
        const formattedExpiresOn = dateFormatter.format(new Date(expiresOn));
        const formattedExtExpiresOn = dateFormatter.format(new Date(extExpiresOn));

        console.log(`\n${BLUE}TOKEN DETAILS${RESET}`);
        console.log(`${BLUE}  Authority${RESET}: ${GREEN}${authority}${RESET}`);
        console.log(`${BLUE}  Scopes${RESET}: ${GREEN}${scopes}${RESET}`);
        console.log(`${BLUE}  Access Token${RESET}: ${GREEN}${accessToken}${RESET}`);

        if (fromCache === 'false') {
            console.log(`${BLUE}  From Cache${RESET}: ${RED}${fromCache}${RESET}`);
        } else {
            console.log(`${BLUE}  From Cache${RESET}: ${GREEN}${fromCache}${RESET}`);
        }

        console.log(`${BLUE}  Expires On${RESET}: ${GREEN}${formattedExpiresOn}${RESET}`);
        console.log(`${BLUE}  Extended Expires On${RESET}: ${GREEN}${formattedExtExpiresOn}${RESET}`);              

        // Wait for 5 seconds (5,000 milliseconds) before making the next call
        await new Promise(resolve => setTimeout(resolve, 5000));
    }
}

main().catch(error => {
    console.error(`${RED}Error in the main function: ${error}${RESET}`);
});

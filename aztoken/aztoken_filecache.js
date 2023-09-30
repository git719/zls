// aztoken_filecache.js
// Node JS version

// ==== House keeping ====================================================
const msal = require('@azure/msal-node');

// Create a token cache instance with persistence
const os = require('os');
const path = require('path');
const fs = require('fs');
const cachePath = path.join(os.homedir(), '.aztoken_cache.json');
fs.closeSync(fs.openSync(cachePath, 'w')); // Touch/create it
const tokenCache = new msal.TokenCache();
// Use persistence options to store cache in a file
const cachePlugin = msal.cachePersistencePlugin(msal.TokenCachePersistenceOptions.fileCache(cachePath));
// Register the plugin with the token cache
tokenCache.registerPersistence(cachePlugin);


// ==== Functions ========================================================
async function getTokenByCredentials(scopes, clientId, clientSecret, authorityUrl) {
  const cca = new msal.ConfidentialClientApplication({
    auth: {
      clientId: clientId,
      authority: authorityUrl,
      clientSecret: clientSecret
    },
    cache: {
      cachePlugin: tokenCache, // Use the configured token cache
    }
    // To not use file caching, you can simply remove above 3 lines and that comma
  });

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


// ==== Main =============================================================
// const scopes = ['https://graph.microsoft.com/.default'];
// const scopes = ['https://management.azure.com/.default'];
const scopes = ['https://ossrdbms-aad.database.windows.net/.default'];
const clientId = process.env.MAZ_CLIENT_ID;
const clientSecret = process.env.MAZ_CLIENT_SECRET;
const authorityUrl = `https://login.microsoftonline.com/${process.env.MAZ_TENANT_ID}`;

getTokenByCredentials(scopes, clientId, clientSecret, authorityUrl)
  .then(token => {
    console.log('Access token:', token);
  })
  .catch(err => {
    console.error('Failed to obtain token:', err);
  });





const { ConfidentialClientApplication, LogLevel } = require('@azure/msal-node');

const clientId = process.env.MAZ_CLIENT_ID;
const clientSecret = process.env.MAZ_CLIENT_SECRET;
const authorityUrl = `https://login.microsoftonline.com/${process.env.MAZ_TENANT_ID}`;

const msalConfig = {
  auth: {
    clientId: clientId,
    authority: authorityUrl,
    clientSecret: clientSecret
  },
  system: {
    loggerOptions: {
      loggerCallback(loglevel, message, containsPii) {
        console.log(message);
      },
      piiLoggingEnabled: false,
      logLevel: LogLevel.Verbose,
    }
  }
};

const cca = new ConfidentialClientApplication(msalConfig);

const scopes = ['https://ossrdbms-aad.database.windows.net/.default'];

async function getTokenByCredentials(scopes) {
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

getTokenByCredentials(scopes)
  .then(token => {
    console.log('Access token:', token);
  })
  .catch(err => {
    console.error('Failed to obtain token:', err);
  });

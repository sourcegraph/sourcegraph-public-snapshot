import fetch from 'node-fetch'

import { version } from '../package.json'

// This will be run after an user has restarted their VSCode after uninstalled
fetch('https://sourcegraph.com/.api/graphql', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        query: `
        mutation {
          logEvent(event: "IDEUninstalled", userCookieID: "VSCE", url: "", source: IDEEXTENSION) {
            alwaysNil
          }
        }
      `,
        variables: { platform: 'vscode-web', version, action: 'uninstalled' },
    }),
})
    .then((response: any) => response.json())
    .then((result: any) => console.log(result))
    .catch(error => console.error(error))

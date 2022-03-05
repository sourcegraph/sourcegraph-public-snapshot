import fetch from 'node-fetch'

import { version } from '../package.json'

/**
 * Script to log extension uninstall event
 * This will be run after an user has uninstalled the extension
 * AND restarted their VSCode
 **/

fetch('https://sourcegraph.com/.api/graphql', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
    },
    body: JSON.stringify({
        query: `
        mutation logEvent($userCookieID: String!, $arguments: String) {
            logEvent(
              event: "IDEUninstalled"
              userCookieID: $userCookieID
              url: ""
              source: BACKEND
              argument: $arguments
              publicArgument: $arguments
            ) {
              alwaysNil
            }
          }
      `,
        variables: {
            userCookieID: '',
            arguments: JSON.stringify({ editor: 'vscode', version }),
        },
    }),
})
    .then((response: any) => response.json())
    .then((result: any) => console.log(result))
    .catch(error => console.error(error))

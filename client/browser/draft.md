Content script

1. client/browser/src/browser-extension/scripts/contentPage.main.ts
- initiates injecting code intel to the code host
2. client/browser/src/shared/code-hosts/shared/inject.ts 
- determine code host
- observe DOM mutations
- inject code intel to the code host
3. client/browser/src/shared/code-hosts/shared/codeHost.tsx
   - client/browser/src/shared/code-hosts/shared/extensions.tsx initialize extensions
     - create platform context (which has `createExtensionHost` method)
       - client/browser/src/shared/platform/context.ts - defines extension host
         - client/browser/src/shared/platform/extensionHost.ts
           - connect to the background thread and create `{endpoints: {proxy, expose}` pair
             - each endpoint (proxy, expose) is a MessagePort
               - MessagePort is created by calling `browserPortToMessagePort` with the `browser.runtime.port` and endpoint role (proxy, expose) as arguments
                 - Links a `browser.runtime.Port` to a MessagePort so it can be used with Comlink across browser extension contexts
                   - stores connected browser ports and callbacks waiting for ports
                   - listens for incoming browser port connections matching a prefix and stores them; calls any waiting callbacks or stores the ports
                   - defines a whenConnected function to call a callback when a port with a given ID connects
                   - defines a link function to set up bidirectional listeners between browser ports and MessagePorts; opens new browser ports for transferred MessagePorts; closes the ports on cleanup
                   - finds MessagePort references in messages from the MessagePort and opens browser ports for them
                   - wraps the message for the browser port to include the port IDs; handles release messages by cleaning up ports
                   - replaces MessagePort references in messages from the browser port with MessagePorts to be transferred; links the browser ports for the IDs when they connect
                   - forwards messages between the ports, transferring any MessagePorts
                   - starts the MessagePort and returns the linked MessagePort
                   
                   - All this allows Comlink to communicate across browser extension contexts by replacing MessagePorts with browser ports and IDs, and re-establishing the MessagePort links when the ports connect.
     - pass platform context to `createExtensionsController` function 
       -  client/shared/src/extensions/createSyncLoadedController.ts 
         - client/shared/src/api/client/connection.ts - create extension host client connection (using the endpoints pair created by calling platform context `createExtensionsController` method)
           -  wait for `createExtensionsController` to resolve endpoints pair
           - wrap (with comlink) `endpoints.proxy` to get `ExtensionHostAPIFactory` from the background script
           - create extension host instance using `ExtensionHostAPIFactory` (FAILS HERE!)
         - returns extension host API 

Background script
1. client/browser/src/browser-extension/scripts/backgroundPage.main.ts
  - listens to `browser.runtime.onInstalled` event and shows after install page when it's fired
  - listens to `browser.tabs.onUpdated` event, executes `js/inject.bundle.js` script at the corresponding tab (if has permissions)
  - defines background page API handlers (`requestGraphQL`, `fetchCache`, `openOptionsPage` and other methods)
  - listens to `browser.runtime.onMessage` event and calls the matching to message type background page API handler passing the message payload as arguments
  - calls `browser.runtime.setUninstallURL` with proper URL as argument
  - creates browser port pairs observable from `browser.runtime.onConnect` event data (coming from the content script)
  - subscribes to browser port pairs changes and handles them
    - creates an extension host worker and client endpoints by calling `createExtensionHostWorker` with `workerBundleURL` (client/shared/src/api/extension/worker.ts)
      - imports client/shared/src/api/extension/main.worker.ts
        - defines and calls `extensionHostMain` function
          - listens to 'message' event (INIT MESSAGE) and extracts endpoint pair from event data (see notes on SENDS INIT EVENT below)
          - starts extension host by calling `startExtensionHost` with this endpoint pair (client/shared/src/api/extension/extensionHost.ts)
            - creates ExtensionHostAPIFactory and exposes it using `comlink.expose` on `endpoints.expose` endpoint (the result of a factory function call IS NEVER RESOLVED IN THE CONTENT SCRIPT)
    
      - creates `clientAPIChannel` MessageChannel 
      - creates `extensionHostAPIChannel` MessageChannel
      - creates worker with script URL set to `workerBundleURL`
      - creates worker endpoints pair (`workerEndpoints`) set to `{ proxy: clientAPIChannel.port2, expose: extensionHostAPIChannel.port2 }`
      - calls `postMessage` on worker (SENDS INIT MESSAGE) with `{ endpoints: workerEndpoints }` as a message and  (`[clientAPIChannel.port2, extensionHostAPIChannel.port2]`) as transfer arguments
      - creates client endpoints pair (`clientEndpoints`) set to `{ proxy: extensionHostAPIChannel.port1, expose: clientAPIChannel.port1 }`
      - returns worker and client endpoints pair
    - links `proxy` and `expose` client endpoints; for each endpoint it
      - links browser port to MessagePort: `proxy` browser port to `proxy` client endpoint (which is a MessagePort) by calling `browserPortToMessagePort` (client/browser/src/shared/platform/ports.ts)
        with the browser port and role (proxy/expose) as arguments (see detailed explanation above - in the content script section)
      - forwards messages between 2 MessagePorts: the client endpoint and the converted to a MessagePort corresponding browser port (proxy-to-proxy, expose-to-expose)
    

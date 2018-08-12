import { BehaviorSubject } from 'rxjs'
import { filter } from 'rxjs/operators'
import WebSocket from 'ws'
import { Client, ClientState } from '../../src/client/client'
import { TextDocumentHoverFeature } from '../../src/client/features/hover'
import { TextDocumentDidOpenFeature } from '../../src/client/features/textDocument'
import { createObservableEnvironment, EMPTY_ENVIRONMENT, Environment } from '../../src/environment/environment'
import { NoopProviderRegistry } from '../../src/environment/providers/registry'
import { createWebSocketMessageTransports } from '../../src/jsonrpc2/transports/nodeWebSocket'
import config from './config'

const environment = new BehaviorSubject<Environment>(EMPTY_ENVIRONMENT)

const client = new Client('', '', {
    root: config.root,
    documentSelector: config.documentSelector,
    initializationOptions: config.initializationOptions,
    createMessageTransports: () =>
        createWebSocketMessageTransports(
            new WebSocket(config.url, {
                headers: { Authorization: `token ${config.accessToken}` },
            })
        ),
})
client.registerFeature(new TextDocumentDidOpenFeature(client, createObservableEnvironment(environment).textDocument))
client.registerFeature(new TextDocumentHoverFeature(client, new NoopProviderRegistry()))
client.state.subscribe(state => console.log('Client state:', ClientState[state]))
client.activate()
const onActive = client.state.pipe(filter(state => state === ClientState.Active))

async function run(): Promise<void> {
    console.log('textDocument/hover...')
    try {
        const result = await client.sendRequest<any>('textDocument/hover', {
            textDocument: { uri: `${config.root}#mux.go` },
            position: { character: 5, line: 23 },
        })
        console.log('textDocument/hover result:', result)
    } catch (err) {
        console.error('textDocument/hover failed:', err.message)
    }

    console.log('textDocument/didOpen...')
    environment.next({
        ...environment.value,
        component: {
            document: {
                uri: `${config.root}#mux.go`,
                languageId: 'go',
                version: 1,
                text: '',
            },
            selections: [],
            visibleRanges: [],
        },
    })
}

onActive.subscribe(async () => {
    await run()
    await client.stop()
    process.exit(0)
})

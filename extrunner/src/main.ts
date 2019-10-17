import puppeteer, { Browser, JSHandle } from 'puppeteer'
import { PlatformContext, StartableEndpoint } from '../../shared/src/platform/context'
import { createExtensionHostClientConnection } from '../../shared/src/api/client/connection'
import { EventEmitter } from 'events'
import { noop } from 'lodash'
import { Services } from '../../shared/src/api/client/services'
import { of, from, Subject } from 'rxjs'
import got from 'got'
import { Settings, SettingsCascadeOrError } from '../../shared/src/settings/settings'
import { take, concatMap, concatAll } from 'rxjs/operators'
import { readFile } from 'mz/fs'
import express from 'express'

interface Dispatcher {
    dispatch(data: any): void
}

async function main(): Promise<void> {
    const app = express()
    app.get('/index.html', (req, res) =>
        res.send(`
            <html>
                <head>
                </head>
            </html>
        `)
    )
    app.use(express.static(__dirname + '/../../../dist/'))
    app.listen(9090)

    const browser = await puppeteer.launch({ headless: false })

    // Set up WebSocket server
    // const httpServer = http.createServer()
    // const webSocketServer = new Server({ server: httpServer })
    // webSocketServer.on('connection', webSocket => {})

    await runExtension('danielevan/sourcegraph-hey-everyone', browser)
}

async function runExtension(extension: string, browser: Browser): Promise<void> {
    // Run extension
    console.log('opening page')
    const page = await browser.newPage()
    await page.goto('http://localhost:9090/index.html')

    await new Promise(resolve => setTimeout(resolve, 3000))

    page.on('console', message => {
        console.log(message.type().toUpperCase(), message.text())
    })

    const createEndpoint = (name: 'proxy' | 'expose'): StartableEndpoint & Dispatcher => {
        const emitter = new EventEmitter()
        return {
            dispatch: (data: unknown): void => {
                emitter.emit('message', { data })
            },
            start: noop,
            postMessage: data => {
                console.log('posting data to', name, data)
                // eslint-disable-next-line @typescript-eslint/no-floating-promises
                page.evaluate(
                    (data, name) => {
                        self.onPuppeteerMessage({ recipient: name, data })
                    },
                    data,
                    name
                )
            },
            addEventListener: (type, listener) => {
                emitter.addListener(type, listener)
            },
            removeEventListener: (type, listener) => {
                emitter.removeListener(type, listener)
            },
        }
    }
    const endpoints = {
        proxy: createEndpoint('expose'),
        expose: createEndpoint('proxy'),
    }
    const postToPuppeteer: typeof self.postToPuppeteer = ({ recipient, data }) => {
        console.log('dispatching message to', recipient, data)
        if (!['proxy', 'expose'].includes(recipient)) {
            throw new Error(`Invalid recipient "${recipient}"`)
        }
        endpoints[recipient].dispatch(data)
    }
    await page.exposeFunction('postToPuppeteer', postToPuppeteer)
    // TODO: this needs to be bundled (or native ESM?)
    await page.addScriptTag({ url: 'extensionHost.bundle.js' })

    const clientApplication = 'other'
    const sourcegraphURL = 'http://localhost:3080'

    const requestGraphQL: PlatformContext['requestGraphQL'] = ({ request, variables }) =>
        from(
            got
                .post(sourcegraphURL + '/.api/graphql', { json: true, body: { request, variables } })
                .then(res => res.body)
        )

    const settings: Settings = {
        extensions: {
            [extension]: true,
        },
    }

    const services = new Services({
        clientApplication,
        getScriptURLForExtension: bundleURL => bundleURL,
        settings: of<SettingsCascadeOrError<Settings>>({
            subjects: [
                {
                    subject: {
                        __typename: 'Client',
                        id: 'clientSettings',
                        displayName: 'Extension Runner Client Settings',
                        viewerCanAdminister: false,
                    },
                    lastID: 1,
                    settings,
                },
            ],
            final: settings,
        }),
        sideloadedExtensionURL: new Subject(),
        requestGraphQL,
        updateSettings: () => Promise.reject(new Error('Cannot update settings in the backend extension runner')),
    })
    console.log('connecting')
    const connection = await createExtensionHostClientConnection(endpoints, services, {
        clientApplication,
        sourcegraphURL,
    })
    console.log('sending ping')
    console.log(await connection.proxy.ping())

    const uri = 'git://github.com/sourcegraph/sourcegraph#test.ts'
    services.model.addModel({
        languageId: 'typescript',
        text: 'Hello',
        uri,
    })

    services.editor.addEditor({
        resource: uri,
        isActive: true,
        selections: [],
        type: 'CodeEditor',
    })

    await from(services.extensions.activeExtensions)
        .pipe(
            concatAll(),
            take(1),
            concatMap(extension => connection.proxy.extensions.$activateExtension(extension.id, extension.scriptURL))
        )
        .toPromise()

    services.textDocumentHover
        .getHover({ textDocument: { uri }, position: { line: 0, character: 0 } })
        .subscribe(console.log)
}

main().catch(err => {
    console.error(err)
    process.exit(1)
})

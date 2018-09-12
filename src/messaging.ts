import { fromEvent, Subscribable, Subscription } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { createMessageConnection, MessageConnection } from 'sourcegraph/module/jsonrpc2/connection'
import { Message } from 'sourcegraph/module/jsonrpc2/messages'
import {
    AbstractMessageReader,
    AbstractMessageWriter,
    DataCallback,
    MessageReader,
    MessageWriter,
} from 'sourcegraph/module/jsonrpc2/transport'
import { UpdateExtensionSettingsArgs } from './context'

class SubscribableMessageReader extends AbstractMessageReader implements MessageReader {
    private pending: Message[] = []
    private callback: DataCallback | null = null
    private subscription = new Subscription()

    constructor(subscribable: Subscribable<Message>) {
        super()

        this.subscription.add(
            subscribable.subscribe(message => {
                try {
                    if (this.callback) {
                        this.callback(message)
                    } else {
                        this.pending.push(message)
                    }
                } catch (err) {
                    this.fireError(err)
                }
            })
        )
    }

    public listen(callback: DataCallback): void {
        if (this.callback) {
            throw new Error('callback is already set')
        }
        this.callback = callback
        while (this.pending.length !== 0) {
            callback(this.pending.pop()!)
        }
    }

    public unsubscribe(): void {
        super.unsubscribe()
        this.subscription.unsubscribe()
    }
}

class CallbackMessageWriter extends AbstractMessageWriter implements MessageWriter {
    constructor(private callback: (message: Message) => void) {
        super()
    }

    public write(message: Message): void {
        this.callback(message)
    }

    public unsubscribe(): void {
        super.unsubscribe()
    }
}

/** One side of a page/client connection. */
export type Source = 'Page' | 'Client'

/**
 * Connects the Sourcegraph extension registry page to a client (such as a browser extension) or vice versa.
 */
function connectAs(source: Source): Promise<MessageConnection> {
    const messageConnection = createMessageConnection({
        reader: new SubscribableMessageReader(
            fromEvent<MessageEvent>(window, 'message').pipe(
                // Filter to relevant messages, ignoring our own
                filter(m => m.data && m.data.source && m.data.source !== source && m.data.message),
                map(m => m.data.message)
            )
        ),
        writer: new CallbackMessageWriter(message => {
            window.postMessage({ source, message }, '*')
        }),
    })

    messageConnection.listen()

    return new Promise(resolve => {
        messageConnection.onNotification('Ping', () => {
            messageConnection.sendNotification('Pong')
            resolve(messageConnection)
        })
        messageConnection.onNotification('Pong', () => {
            resolve(messageConnection)
        })
        messageConnection.sendNotification('Ping')
    })
}

/** A connection to the client. */
export interface ClientConnection {
    /** Listens for the latest client settings. */
    onSettings: (callback: (settings: string) => void) => void

    /** Requests the client to update a setting. */
    editSetting: (edit: UpdateExtensionSettingsArgs) => Promise<void>

    /** Asks the client for its settings. */
    getSettings: () => Promise<string>

    /** The underlying JSON RPC connection. */
    rawConnection: MessageConnection
}

/** A connection to the page. */
export interface PageConnection {
    /** Listens for requests to edit settings. */
    onEditSetting: (callback: (edit: UpdateExtensionSettingsArgs) => Promise<void>) => void

    /** Listens for requests for the latest settings. */
    onGetSettings: (callback: () => Promise<string>) => void

    /** Notifies the page that the settings have been updated. */
    sendSettings: (settings: string) => void

    /** The underlying JSON RPC connection. */
    rawConnection: MessageConnection
}

/**
 * Connects the client (such as a browser extension) to a Sourcegraph extension registry page.
 */
export function connectAsClient(): Promise<PageConnection> {
    return connectAs('Client').then<PageConnection>(connection => ({
        onEditSetting: callback => {
            connection.onRequest('EditSetting', callback)
        },
        onGetSettings: callback => {
            connection.onRequest('GetSettings', callback)
        },
        sendSettings: settings => {
            connection.sendNotification('Settings', settings)
        },
        rawConnection: connection,
    }))
}

/**
 * Connects the Sourcegraph extension registry page to a client (such as a browser extension).
 */
export function connectAsPage(): Promise<ClientConnection> {
    return connectAs('Page').then<ClientConnection>(connection => ({
        onSettings: callback => {
            connection.onNotification('Settings', callback)
        },
        editSetting: callback => connection.sendRequest('EditSetting', callback),
        getSettings: () => connection.sendRequest('GetSettings'),
        rawConnection: connection,
    }))
}

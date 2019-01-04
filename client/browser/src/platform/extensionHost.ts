import { Observable, of } from 'rxjs'
import uuid from 'uuid'
import { MessageTransports } from '../../../../shared/src/api/protocol/jsonrpc2/connection'
import { Message } from '../../../../shared/src/api/protocol/jsonrpc2/messages'
import {
    AbstractMessageReader,
    AbstractMessageWriter,
    DataCallback,
    MessageReader,
    MessageWriter,
} from '../../../../shared/src/api/protocol/jsonrpc2/transport'
import { createWebWorkerMessageTransports } from '../../../../shared/src/api/protocol/jsonrpc2/transports/webWorker'
import { isInPage } from '../context'
import { createExtensionHostWorker } from './worker'

/**
 * Spawns an extension and returns a communication channel to it.
 */
export function createExtensionHost(): Observable<MessageTransports> {
    if (isInPage) {
        return createInPageExtensionHost()
    }
    const channelID = uuid.v4()
    return of(createPortMessageTransports(chrome.runtime.connect({ name: channelID })))
}

function createInPageExtensionHost(): Observable<MessageTransports> {
    const worker = createExtensionHostWorker()
    const messageTransports = createWebWorkerMessageTransports(worker)
    return new Observable(sub => {
        sub.next(messageTransports)
        return () => worker.terminate()
    })
}

class PortMessageReader extends AbstractMessageReader implements MessageReader {
    private pending: Message[] = []
    private callback: DataCallback | null = null

    constructor(private port: chrome.runtime.Port) {
        super()

        port.onMessage.addListener((message: any) => {
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
        port.onDisconnect.addListener(() => {
            this.fireClose()
        })
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
        this.callback = null
        this.port.disconnect()
    }
}

class PortMessageWriter extends AbstractMessageWriter implements MessageWriter {
    private errorCount = 0

    constructor(private port: chrome.runtime.Port) {
        super()
    }

    public write(message: Message): void {
        try {
            this.port.postMessage(message)
        } catch (error) {
            this.fireError(error, message, ++this.errorCount)
        }
    }

    public unsubscribe(): void {
        super.unsubscribe()
        this.port.disconnect()
    }
}

/** Creates JSON-RPC2 message transports for the Web Worker message communication interface. */
function createPortMessageTransports(port: chrome.runtime.Port): MessageTransports {
    return {
        reader: new PortMessageReader(port),
        writer: new PortMessageWriter(port),
    }
}

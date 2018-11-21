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
import { ConfiguredExtension } from '../../../../shared/src/extensions/extension'
import { SettingsCascade } from '../../../../shared/src/settings/settings'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { ExtensionConnectionInfo, onFirstMessage } from '../messaging'

/**
 * Spawns an extension and returns a communication channel to it.
 */
export function createMessageTransports(
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>,
    settingsCascade: SettingsCascade
): Promise<MessageTransports> {
    if (!extension.manifest) {
        throw new Error(`unable to connect to extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    const manifest = extension.manifest
    if (isErrorLike(manifest)) {
        throw new Error(
            `unable to connect to extension ${JSON.stringify(extension.id)}: invalid manifest: ${manifest.message}`
        )
    }
    return new Promise<MessageTransports>((resolve, reject) => {
        const channelID = uuid.v4()
        const port = chrome.runtime.connect({ name: channelID })
        port.postMessage({
            extensionID: extension.id,
            jsBundleURL: manifest.url,
            settingsCascade,
        } as ExtensionConnectionInfo)
        onFirstMessage(port, (response: { error?: any }) => {
            if (response.error) {
                reject(response.error)
            } else {
                resolve(createPortMessageTransports(port))
            }
        })
    }).catch(err => {
        console.error('Error connecting to', extension.id + ':', err)
        throw err
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

import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import { ClientOptions } from 'cxp/module/client/client'
import { MessageTransports } from 'cxp/module/jsonrpc2/connection'
import uuid from 'uuid'
import { onFirstMessage } from '../../chrome/extension/background'
import { ExtensionConnectionInfo } from '../../chrome/extension/background'
import { isErrorLike } from './errors'
import { createPortMessageTransports } from './PortMessageTransports'

const createPlatformMessageTransports = (connectionInfo: ExtensionConnectionInfo) =>
    new Promise<MessageTransports>((resolve, reject) => {
        const channelID = uuid.v4()
        const port = chrome.runtime.connect({ name: channelID })
        port.postMessage(connectionInfo)
        onFirstMessage(port, (response: { error?: any }) => {
            if (response.error) {
                reject(response.error)
            } else {
                resolve(createPortMessageTransports(port, connectionInfo))
            }
        })
    })

export function createMessageTransports(
    extension: Pick<ConfiguredExtension, 'id' | 'manifest'>,
    options: ClientOptions
): Promise<MessageTransports> {
    if (!extension.manifest) {
        throw new Error(`unable to connect to extension ${JSON.stringify(extension.id)}: no manifest found`)
    }
    if (!options.root) {
        throw new Error(`unable to connect to extension ${JSON.stringify(extension.id)}: no root`)
    }
    if (isErrorLike(extension.manifest)) {
        throw new Error(
            `unable to connect to extension ${JSON.stringify(extension.id)}: invalid manifest: ${
                extension.manifest.message
            }`
        )
    }
    if (['websocket', 'tcp', 'bundle'].includes(extension.manifest.platform.type)) {
        return createPlatformMessageTransports({
            extensionID: extension.id,
            platform: extension.manifest.platform,
            rootURI: options.root,
        }).catch(err => {
            console.error('Error connecting to', extension.id + ':', err)
            throw err
        })
    } else {
        return Promise.reject(
            new Error(
                `Unable to connect to CXP extension ${JSON.stringify(extension.id)}: type ${JSON.stringify(
                    extension.manifest.platform.type
                )} is not supported`
            )
        )
    }
}

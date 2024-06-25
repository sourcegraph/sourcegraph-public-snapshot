import * as vscode from 'vscode'

import { extensionContext } from '../extension'

const defaultEndpointURL = 'https://sourcegraph.com'

const endpointKey = 'sourcegraph.url'

async function removeOldEndpointURLSetting(): Promise<void> {
    await vscode.workspace.getConfiguration().update(endpointKey, undefined, vscode.ConfigurationTarget.Global)
    await vscode.workspace.getConfiguration().update(endpointKey, undefined, vscode.ConfigurationTarget.Workspace)
    return
}

export function endpointSetting(): string {
    // get the URl from either, 1. extension local storage (new)
    let url = extensionContext?.globalState.get<string>(endpointKey)
    if (!url) {
        // 2. settings.json (old)
        url = vscode.workspace.getConfiguration().get<string>(endpointKey)
        if (url) {
            // if settings.json, migrate to extension local storage
            extensionContext?.globalState.update(endpointKey, url).then(
                () => {
                    void removeOldEndpointURLSetting()
                },
                error => {
                    console.error(error)
                }
            )
        } else {
            // or, 3. default value
            url = defaultEndpointURL
        }
    }
    return removeEndingSlash(url)
}

export function setEndpoint(newEndpoint: string | undefined): void {
    const newEndpointURL = newEndpoint ? removeEndingSlash(newEndpoint) : defaultEndpointURL
    const currentEndpointHostname = new URL(endpointSetting()).hostname
    const newEndpointHostname = new URL(newEndpointURL).hostname
    if (currentEndpointHostname !== newEndpointHostname) {
        extensionContext?.globalState.update(endpointKey, newEndpointURL).then(
            () => {},
            error => {
                console.error(error)
            }
        )
    }
}

export function endpointHostnameSetting(): string {
    return new URL(endpointSetting()).hostname
}

export function endpointPortSetting(): number {
    const port = new URL(endpointSetting()).port
    return port ? parseInt(port, 10) : 443
}

export function endpointProtocolSetting(): string {
    return new URL(endpointSetting()).protocol
}

export function endpointRequestHeadersSetting(): object {
    return vscode.workspace.getConfiguration().get<object>('sourcegraph.requestHeaders') || {}
}

function removeEndingSlash(uri: string): string {
    if (uri.endsWith('/')) {
        return uri.slice(0, -1)
    }
    return uri
}

export function isSourcegraphDotCom(): boolean {
    const hostname = new URL(endpointSetting()).hostname
    return hostname === 'sourcegraph.com' || hostname === 'www.sourcegraph.com'
}

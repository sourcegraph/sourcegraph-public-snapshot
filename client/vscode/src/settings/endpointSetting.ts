import * as vscode from 'vscode'

import { readConfiguration } from './readConfiguration'

export function endpointSetting(): string {
    const url = readConfiguration().get<string>('url') || 'https://sourcegraph.com'
    return removeEndingSlash(url)
}

export async function setEndpoint(newEndpoint: string | undefined): Promise<void> {
    const newEndpointURL = newEndpoint ? removeEndingSlash(newEndpoint) : undefined
    await readConfiguration().update('url', newEndpointURL, vscode.ConfigurationTarget.Global)
    await readConfiguration().update('url', newEndpointURL, vscode.ConfigurationTarget.Workspace)
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
    return readConfiguration().get<object>('requestHeaders') || {}
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

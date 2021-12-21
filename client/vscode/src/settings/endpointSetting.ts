import * as vscode from 'vscode'

import { readConfiguration } from './readConfiguration'

export function activateEndpointSetting(): vscode.Disposable {
    return vscode.workspace.onDidChangeConfiguration(config => {
        if (config.affectsConfiguration('sourcegraph.url')) {
            // TODO reload extension (or invalidate gql if we have to)
        }
    })
}

export function endpointSetting(): string {
    // has default value
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    const url = readConfiguration().get<string>('url')!
    if (url.endsWith('/')) {
        return url.slice(0, -1)
    }
    return url
}

export function endpointHostnameSetting(): string {
    return new URL(endpointSetting()).hostname
}

export function endpointPortSetting(): number {
    const port = new URL(endpointSetting()).port
    return port ? parseInt(port, 10) : 443
}

// Check if Access Token is configured in setting
export function endpointAccessTokenSetting(): boolean {
    if (readConfiguration().get<string>('accessToken')) {
        return true
    }
    return false
}

// Check if Cors is configured in setting
export function endpointCorsSetting(): string {
    const corsUrl = readConfiguration().get<string>('corsUrl')
    if (corsUrl) {
        return new URL('', corsUrl).origin
    }
    return ''
}

// Update Cors in setting
export async function updateCorsSetting(corsUrl: string): Promise<boolean> {
    try {
        const newCorsUrl = new URL('', corsUrl).origin
        await readConfiguration().update('corsUrl', newCorsUrl, vscode.ConfigurationTarget.Global)
        return true
    } catch {
        return false
    }
}

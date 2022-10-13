import { readConfiguration } from './readConfiguration'

let instance_url = 'https://sourcegraph.com'

export function setEndpointSetting(uri: string): void {
    instance_url = uri
}

export function endpointSetting(): string {
    return removeEndingSlash(instance_url)
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

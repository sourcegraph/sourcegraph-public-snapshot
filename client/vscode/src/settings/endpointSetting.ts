import { readConfiguration } from './readConfiguration'

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

export function endpointAccessTokenSetting(): boolean {
    if (readConfiguration().get<string>('accessToken')) {
        return true
    }
    return false
}

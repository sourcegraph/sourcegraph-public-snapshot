import fetch from 'isomorphic-fetch'

import { ConfigurationWithAccessToken } from '../../configuration'

export class SourcegraphRestAPIClient {
    constructor(
        private config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
    ) {}

    async fetch<T = unknown>(path: string, options?: RequestInit): Promise<T> {
        const headers = new Headers(this.config.customHeaders as HeadersInit)

        headers.set('Content-Type', 'application/json; charset=utf-8')

        if (this.config.accessToken) {
            headers.set('Authorization', `token ${this.config.accessToken}`)
        }

        const url = this.makeUrl(path)

        console.log('url', url, headers)

        const response = await fetch(url, {
            ...options,
            headers,
        })

        if (!response.ok) {
            throw new Error(`HTTP error: ${response.status} ${response.statusText}`)
        }

        return response.json() as T
    }

    private makeUrl(path: string): string {
        // const BASE_PATH = this.config.serverEndpoint
        const BASE_PATH = 'https://cody-gateway.sourcegraph.com'

        const pathPart = path.startsWith('/') ? path : '/' + path

        return BASE_PATH + pathPart
    }
}

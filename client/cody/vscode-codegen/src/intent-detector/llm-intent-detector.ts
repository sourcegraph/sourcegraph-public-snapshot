import fetch from 'node-fetch'

import { QueryInfo } from '@sourcegraph/cody-common'

import { IntentDetector } from '.'

export class LLMIntentDetector implements IntentDetector {
    constructor(private serverUrl: string, private accessToken: string) {}

    public async detect(text: string): Promise<QueryInfo> {
        const uri = new URL('/info', this.serverUrl)
        const searchParameters = new URLSearchParams()
        searchParameters.set('q', text)
        uri.search = searchParameters.toString()
        const resp = await fetch(uri.href, {
            method: 'GET',
            headers: {
                Authorization: `Bearer ${this.accessToken}`,
            },
        })
        const respJSON = await resp.json()
        if (!('needsCodebaseContext' in respJSON) || !('needsCurrentFileContext' in respJSON)) {
            throw new Error(`malformed response from /info: ${JSON.stringify(respJSON)}`)
        }
        return respJSON as QueryInfo
    }
}

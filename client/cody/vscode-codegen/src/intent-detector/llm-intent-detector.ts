import fetch from 'node-fetch'

import { QueryInfo } from '@sourcegraph/cody-common'

import { IntentDetector } from '.'

export class LLMIntentDetector implements IntentDetector {
    constructor(private serverUrl: string, private accessToken: string) {}

    public async detect(text: string): Promise<QueryInfo> {
        const resp = await fetch(`${this.serverUrl}/info?q=${encodeURIComponent(text)}`, {
            method: 'GET',
            headers: {
                Authorization: 'Bearer ' + this.accessToken,
            },
        })
        const respJSON = await resp.json()
        if (!('needsCodebaseContext' in respJSON) || !('needsCurrentFileContext' in respJSON)) {
            throw new Error(`malformed response from /info: ${JSON.stringify(respJSON)}`)
        }
        return respJSON as QueryInfo
    }
}

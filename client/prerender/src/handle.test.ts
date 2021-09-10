import fetch from 'node-fetch'

import type { RenderRequest, RenderResponse } from './render'

const port = 3190 // TODO(sqs): dedupe with serve.ts

// Requires the prerender server to be running.
const render = async (request: RenderRequest): Promise<RenderResponse> => {
    const response = await fetch(`http://localhost:${port}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json; charset=utf-8' },
        body: JSON.stringify(request),
    })
    return (await response.json()) as RenderResponse
}

describe('handle', () => {
    test('render', async () => {
        expect((await render({ requestURI: '/search', jscontext: {} })).html).toBe('<!--$!-->Loading app...<!--/$-->')
    })
})

import './browserEnv'

import http from 'http'

import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { render, RenderRequest, RenderResponse } from './render'

const port = process.env.PORT || 3190

const wrappedRender = async (request: RenderRequest): Promise<RenderResponse> => {
    try {
        return await render(request)
    } catch (error) {
        console.error(`Error (${request.requestURI}):`, error)
        return {
            error: isErrorLike(error)
                ? error.name
                    ? `${error.name}: ${error.message}`
                    : error.message
                : String(error),
        }
    }
}

const jsonRequestBody = async (request: http.IncomingMessage): Promise<unknown> => {
    const buffers = []
    for await (const chunk of request) {
        buffers.push(chunk)
    }
    return JSON.parse(Buffer.concat(buffers).toString()) as unknown
}

// TODO(sqs): eslint disable below, how can we remove it?
// eslint-disable-next-line @typescript-eslint/no-misused-promises
http.createServer(async (request, response) => {
    if (request.method !== 'POST') {
        response.writeHead(405)
        response.write('method must be POST\n')
        response.end()
        return
    }

    let renderRequest: RenderRequest
    try {
        renderRequest = (await jsonRequestBody(request)) as RenderRequest
    } catch {
        response.writeHead(400)
        response.write('invalid RenderRequest JSON in request body\n')
        response.end()
        return
    }

    const renderResponse = await wrappedRender(renderRequest)
    response.writeHead(200, { 'Content-Type': 'application/json; charset=utf-8' })
    response.write(JSON.stringify(renderResponse))
    response.write('\n')
    response.end()
}).listen(port)

console.error(`Ready at: http://localhost:${port}`)

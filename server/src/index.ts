import { createServer } from 'http'
import { parse } from 'url'
import { WebSocketServer } from 'ws'
import express from 'express'

import * as bodyParser from 'body-parser'
import {
	WSChatRequest,
	WSChatResponseChange,
	WSChatResponseComplete,
	WSChatResponseError,
} from '@sourcegraph/cody-common'
import { ClaudeBackend } from './prompts/claude'
import { wsHandleGetCompletions } from './completions'
import { authenticate, getUsers } from './auth'

const anthropicApiKey = process.env.ANTHROPIC_API_KEY
if (!anthropicApiKey) {
	throw new Error('ANTHROPIC_API_KEY is missing')
}

const usersPath = process.env.CODY_USERS_PATH
if (!usersPath) {
	throw new Error('CODY_USERS_PATH is missing')
}

const port = process.env.CODY_PORT || '8080'

const claudeBackend = new ClaudeBackend(anthropicApiKey, {
	model: 'claude-v1',
	temperature: 0.2,
	stop_sequences: ['\n\nHuman:'],
	max_tokens_to_sample: 1000,
	top_p: 1.0,
	top_k: -1,
})

const app = express()
app.use(bodyParser.json())

const httpServer = createServer(app)

const wssCompletions = new WebSocketServer({ noServer: true })
wssCompletions.on('connection', ws => {
	console.log('completions:connection')
	ws.on('message', async data => {
		try {
			console.log('completions:request')
			const req = JSON.parse(data.toString())
			switch (req.kind) {
				case 'getCompletions':
					if (req.kind !== 'getCompletions' || !req.args || !req.requestId) {
						console.error(`invalid request ${data.toString()}`)
						return
					}
					await wsHandleGetCompletions(ws, req)
					return
				default:
					console.error(`invalid request ${data.toString()}`)
					return
			}
		} catch (error: any) {
			console.error('Uncaught error', error)
		}
	})
	// TODO(beyang): handle shutdown
})

const wssChat = new WebSocketServer({ noServer: true })
wssChat.on('connection', ws => {
	console.log('chat:connection')
	// TODO(beyang): Close connection after timeout. Probably should keep connection around,
	// rather than closing after every response?

	ws.on('message', async data => {
		console.log('chat:request')
		const req = JSON.parse(data.toString()) as WSChatRequest
		if (!req.requestId || !req.messages) {
			console.error(`invalid request ${data.toString()}`)
			return
		}
		claudeBackend.chat(req.messages, {
			onChange: message => {
				const msg: WSChatResponseChange = { requestId: req.requestId, kind: 'response:change', message }
				ws.send(JSON.stringify(msg))
			},
			onComplete: message => {
				const msg: WSChatResponseComplete = { requestId: req.requestId, kind: 'response:complete', message }
				ws.send(JSON.stringify(msg), err => {
					if (err) {
						console.error(`error sending last response message: ${err}`)
					}
				})
			},
			onError: error => {
				const msg: WSChatResponseError = { requestId: req.requestId, kind: 'response:error', error }
				ws.send(JSON.stringify(msg), err => {
					if (err) {
						console.error(`error sending error message: ${err}`)
					}
				})
			},
		})
	})
})

httpServer.on('upgrade', (request, socket, head) => {
	if (!request.url) {
		return
	}

	const { pathname, search } = parse(request.url)

	const user = authenticate(
		request.headers['authorization'],
		new URLSearchParams(search || '').get('access_token'),
		getUsers(usersPath)
	)
	if (!user) {
		socket.end('HTTP/1.1 401 Unauthorized\r\n\r\n')
		return
	}

	if (pathname === '/completions') {
		wssCompletions.handleUpgrade(request, socket, head, ws => {
			wssCompletions.emit('connection', ws, request)
		})
	} else if (pathname === '/chat') {
		wssChat.handleUpgrade(request, socket, head, ws => {
			wssChat.emit('connection', ws, request)
		})
	} else {
		socket.destroy()
	}
})

console.log(`Server listening on :${port}`)
httpServer.listen(port)

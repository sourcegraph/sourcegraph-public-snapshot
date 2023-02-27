import * as vscode from 'vscode'
import { WebSocket } from 'ws'

import { WSResponse } from '@sourcegraph/cody-common'

export class WSClient<TRequest, TResponse extends WSResponse> {
	static async new<T1, T2 extends WSResponse>(addr: string, accessToken: string): Promise<WSClient<T1, T2> | null> {
		try {
			const options = { headers: { authorization: `Bearer ${accessToken}` } }
			const ws = new WebSocket(httpToWSURL(addr), options)
			const c = new WSClient<T1, T2>(ws, options)
			await c.waitForConnection(30 * 1000) // 30 seconds
			return c
		} catch (error) {
			vscode.window.showWarningMessage(
				'Could not connect to the Cody backend. Check that you have set the correct access token.'
			)
			console.error(error)
			return null
		}
	}

	private nextRequestId = 1
	private readonly responseListeners: {
		[id: number]: (resp: TResponse) => boolean
	} = {}

	constructor(private ws: WebSocket, private options: { headers: { authorization: string } }) {
		this.addHandlers()
	}

	private addHandlers() {
		this.ws.on('message', rawMsg => {
			const msg: TResponse = JSON.parse(rawMsg.toString())
			if (!msg.requestId) {
				return
			}
			const handler = this.responseListeners[msg.requestId]
			if (!handler) {
				return
			}
			const isLastResponse = handler(msg)
			if (isLastResponse) {
				delete this.responseListeners[msg.requestId]
			}
		})
		this.ws.on('error', err => {
			console.error(`websocket error: ${err}`)
		})
	}

	async ensureConnected(): Promise<void> {
		const readyState = this.ws.readyState
		switch (readyState) {
			case WebSocket.OPEN:
				return
			case WebSocket.CONNECTING:
				await this.waitForConnection(30 * 1000)
				return
			case WebSocket.CLOSED:
			case WebSocket.CLOSING:
				console.log(`reconnecting to ${this.ws.url}`)
				this.ws = new WebSocket(this.ws.url, this.options)
				this.addHandlers()
				await this.waitForConnection(30 * 1000)
				return
			default:
				throw new Error(`unrecognized websocket ready state: ${readyState}`)
		}
	}

	private async waitForConnection(openTimeout: number): Promise<void> {
		await Promise.race([
			new Promise<void>(resolve =>
				this.ws.on('open', () => {
					resolve()
				})
			),
			new Promise<void>((_, reject) => {
				setTimeout(() => {
					reject(`Failed to create websocket connection, timed out in ${openTimeout}ms`)
				}, openTimeout)
			}),
		])
	}

	async sendRequest(req: TRequest, handleResponse: (resp: TResponse) => boolean): Promise<() => void> {
		const requestId = this.nextRequestId++
		this.responseListeners[requestId] = handleResponse
		const reqWithId = {
			...req,
			requestId,
		}
		await this.ensureConnected()

		this.ws.send(JSON.stringify(reqWithId), async err => {
			if (err) {
				throw new Error(`failed to send websocket request: ${err}`)
			}
		})

		// A callback to close (or ignore the responses for) the current request.
		return () => {
			delete this.responseListeners[requestId]
		}
	}
}

function httpToWSURL(httpURL: string): string {
	const url = new URL(httpURL)
	if (url.protocol !== 'http:' && url.protocol !== 'https:') {
		throw new Error(`URL protocol must be either http: or https: (got ${url.protocol})`)
	}
	url.protocol = url.protocol === 'http:' ? 'ws:' : 'wss:'
	return url.toString()
}

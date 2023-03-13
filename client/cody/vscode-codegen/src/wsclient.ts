import WebSocket from 'isomorphic-ws'
import * as vscode from 'vscode'

import { WSResponse } from '@sourcegraph/cody-common'

export class WSClient<TRequest, TResponse extends WSResponse> {
	public static async new<T1, T2 extends WSResponse>(
		addr: string,
		accessToken: string
	): Promise<WSClient<T1, T2> | null> {
		try {
			const options = { headers: { authorization: `Bearer ${accessToken}` } }
			const socket = new WebSocket(httpToWSURL(addr), options)
			const client = new WSClient<T1, T2>(socket, options)
			await client.waitForConnection(30 * 1000) // 30 seconds
			return client
		} catch (error) {
			void vscode.window.showWarningMessage(
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

	constructor(private socket: WebSocket, private options: { headers: { authorization: string } }) {
		this.addHandlers()
	}

	private addHandlers(): void {
		this.socket.on('message', rawMsg => {
			// eslint-disable-next-line @typescript-eslint/no-base-to-string
			const msg: TResponse = JSON.parse(rawMsg.toString()) as TResponse
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
		this.socket.on('error', err => {
			console.error(`websocket error: ${err.toString()}`)
		})
	}

	private async ensureConnected(): Promise<void> {
		const readyState = this.socket.readyState
		switch (readyState) {
			case WebSocket.OPEN:
				return
			case WebSocket.CONNECTING:
				await this.waitForConnection(30 * 1000)
				return
			case WebSocket.CLOSED:
			case WebSocket.CLOSING:
				console.log(`reconnecting to ${this.socket.url}`)
				this.socket = new WebSocket(this.socket.url, this.options)
				this.addHandlers()
				await this.waitForConnection(30 * 1000)
				return
			default:
				throw new Error(`unreachable websocket ready state: ${readyState as number}`)
		}
	}

	private async waitForConnection(openTimeout: number): Promise<void> {
		return new Promise<void>((resolve, reject) => {
			this.socket.on('open', resolve)
			setTimeout(() => {
				reject(new Error(`Failed to create websocket connection, timed out in ${openTimeout}ms`))
			}, openTimeout)
		})
	}

	public async sendRequest(req: TRequest, handleResponse: (resp: TResponse) => boolean): Promise<() => void> {
		const requestId = this.nextRequestId++
		this.responseListeners[requestId] = handleResponse
		const reqWithId = {
			...req,
			requestId,
		}
		await this.ensureConnected()

		this.socket.send(JSON.stringify(reqWithId), err => {
			if (err) {
				console.error(`failed to send websocket request: ${err.toString()}`)
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

import * as vscode from "vscode";
import { WSResponse } from "common";
import { WebSocket } from "ws";

export class WSClient<TRequest, TResponse extends WSResponse> {
	static async new<T1, T2 extends WSResponse>(
		addr: string
	): Promise<WSClient<T1, T2>> {
		const ws = new WebSocket(addr);
		const c = new WSClient<T1, T2>(ws);
		await c.waitForConnection(30 * 1000); // 30 seconds
		return c;
	}

	private nextRequestId = 1;
	private ws: WebSocket;
	private readonly responseListeners: {
		[id: number]: (resp: TResponse) => boolean;
	} = {};

	constructor(ws: WebSocket) {
		this.ws = ws;
		this.addHandlers();
	}

	private addHandlers() {
		this.ws.on("message", (rawMsg) => {
			const msg: TResponse = JSON.parse(rawMsg.toString());
			if (!msg.requestId) {
				return;
			}
			const handler = this.responseListeners[msg.requestId];
			if (!handler) {
				console.error(`did not find handler for requestId ${msg.requestId}`);
				return;
			}
			const isLastResponse = handler(msg);
			if (isLastResponse) {
				delete this.responseListeners[msg.requestId];
			}
		});
		this.ws.on("error", (err) => {
			vscode.window.showErrorMessage(`websocket error: ${err}`);
		});
	}

	async ensureConnected(): Promise<void> {
		const readyState = this.ws.readyState;
		switch (readyState) {
			case WebSocket.OPEN:
				return;
			case WebSocket.CONNECTING:
				await this.waitForConnection(30 * 1000);
				return;
			case WebSocket.CLOSED:
			case WebSocket.CLOSING:
				console.log(`reconnecting to ${this.ws.url}`);
				this.ws = new WebSocket(this.ws.url);
				this.addHandlers();
				await this.waitForConnection(30 * 1000);
				return;
			default:
				throw new Error(`unrecognized websocket ready state: ${readyState}`);
		}
	}

	private async waitForConnection(openTimeout: number): Promise<void> {
		await Promise.race([
			new Promise<void>((resolve) =>
				this.ws.on("open", () => {
					resolve();
				})
			),
			new Promise<void>((_, reject) => {
				setTimeout(() => {
					reject(
						`Failed to create websocket connection, timed out in ${openTimeout}ms`
					);
				}, openTimeout);
			}),
		]);
	}

	async sendRequest(
		req: TRequest,
		handleResponse: (resp: TResponse) => boolean
	): Promise<void> {
		const requestId = this.nextRequestId++;
		this.responseListeners[requestId] = handleResponse;
		const reqWithId = {
			...req,
			requestId,
		};
		await this.ensureConnected();

		this.ws.send(JSON.stringify(reqWithId), async (err) => {
			if (err) {
				throw new Error(`failed to send websocket request: ${err}`);
			}
		});
	}
}

import * as vscode from "vscode";
import { WSResponse } from "common";
import { WebSocket } from "ws";

export class WSClient<TRequest, TResponse extends WSResponse> {
	static async new<T1, T2 extends WSResponse>(
		addr: string
	): Promise<WSClient<T1, T2>> {
		const ws = new WebSocket(addr);
		const c = new WSClient<T1, T2>(ws);
		const openTimeout = 30 * 1000; // 30 seconds
		await Promise.race([
			new Promise<void>((resolve) =>
				ws.on("open", () => {
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
		return c;
	}

	private nextRequestId = 1;
	private readonly ws: WebSocket;
	private readonly responseListeners: {
		[id: number]: (resp: TResponse) => boolean;
	} = {};

	constructor(ws: WebSocket) {
		ws.on("message", (rawMsg) => {
			const msg: TResponse = JSON.parse(rawMsg.toString());
			if (!msg.requestId) {
				return;
			}
			const handler = this.responseListeners[msg.requestId];
			if (!handler) {
				return;
			}
			const isLastResponse = handler(msg);
			if (isLastResponse) {
				delete this.responseListeners[msg.requestId];
			}
		});
		ws.on("error", (err) => {
			vscode.window.showErrorMessage(`websocket error: ${err}`);
		});
		this.ws = ws;
	}

	sendRequest(req: TRequest, handleResponse: (resp: TResponse) => boolean) {
		const requestId = this.nextRequestId++;
		this.responseListeners[requestId] = handleResponse;
		const reqWithId = {
			...req,
			requestId,
		};
		this.ws.send(JSON.stringify(reqWithId), (err) => {
			if (err) {
				throw new Error(`failed to send websocket request: ${err}`);
			}
		});
	}
}

import { DataCallback, Message, MessageReader, MessageWriter } from "@sourcegraph/vscode-jsonrpc";
import { AbstractMessageReader } from "@sourcegraph/vscode-jsonrpc/lib/messageReader";
import { AbstractMessageWriter } from "@sourcegraph/vscode-jsonrpc/lib/messageWriter";
import { MessageStream } from "@sourcegraph/vscode-languageclient";

// connectWebSocket can be passed to the Zap client's ServerOptions to
// connect to a Zap server via a WebSocket.
export function webSocketStreamOpener(url: string): () => Promise<MessageStream> {
	return () => {
		return new Promise((resolve, reject) => {
			let socket = new WebSocket(url);
			socket.binaryType = "arraybuffer";
			let connected = false;
			socket.onopen = () => {
				connected = true;
				resolve({
					reader: new WebSocketMessageReader(socket),
					writer: new WebSocketMessageWriter(socket),
				});
			};
			socket.onclose = (ev: CloseEvent) => {
				if (ev.code !== 1000 /* Close code: Normal */) {
					console.error("WebSocket closed:", ev);
				}
				if (!connected) {
					reject(ev);
				}
			};
		});
	};
};

/**
 * WebSocketMessageReader wraps a WebSocket to conform to the MessageReader interface.
 */
class WebSocketMessageReader extends AbstractMessageReader implements MessageReader {
	private socket: WebSocket;
	private callback: DataCallback;

	constructor(socket: WebSocket) {
		super();
		this.socket = socket;

		socket.onmessage = (ev: MessageEvent) => {
			if (!this.callback) {
				this.fireError(new Error("message arrived on WebSocket but there is no listener"));
				return;
			}
			try {
				const data = JSON.parse(ev.data);
				this.callback(data);
			} catch (error) {
				this.fireError(error);
			}
		};
	}

	public listen(callback: DataCallback): void {
		this.callback = callback;
	}
}

/**
 * WebSocketMessageWriter wraps a WebSocket to conform to the MessageWriter interface.
 */
class WebSocketMessageWriter extends AbstractMessageWriter implements MessageWriter {
	private socket: WebSocket;
	private socketClosed: boolean;
	private errorCount: number;

	constructor(socket: WebSocket) {
		super();
		this.socket = socket;
		this.socketClosed = false;
		this.errorCount = 0;

		socket.onclose = (ev: CloseEvent) => {
			this.socketClosed = true;
			this.fireClose();
		};
		socket.onerror = (ev: ErrorEvent) => {
			this.fireError(ev.error);
		};
	}

	public write(msg: Message): void {
		if (this.socketClosed) {
			this.errorCount++;
			this.fireError(new Error("Write on closed WebSocket"), msg, this.errorCount);
			return;
		}
		this.errorCount = 0;
		this.socket.send(JSON.stringify(msg));
	}
}

import { DataCallback, Message, MessageReader, MessageWriter, RequestMessage } from "@sourcegraph/vscode-jsonrpc";
import { AbstractMessageReader } from "@sourcegraph/vscode-jsonrpc/lib/messageReader";
import { AbstractMessageWriter } from "@sourcegraph/vscode-jsonrpc/lib/messageWriter";
import { MessageStream } from "@sourcegraph/vscode-languageclient";

// connectWebSocket can be passed to the Zap client's ServerOptions to
// connect to a Zap server via a WebSocket.
export function webSocketStreamOpener(url: string, requestTracer?: (trace: MessageTrace) => void): () => Promise<MessageStream> {
	return () => {
		return new Promise((resolve, reject) => {
			let socket = new WebSocket(url);
			socket.binaryType = "arraybuffer";
			let connected = false;
			socket.onopen = () => {
				connected = true;
				const reader = new WebSocketMessageReader(socket);
				const writer = new WebSocketMessageWriter(socket);
				if (requestTracer) {
					traceJSONRPCRequests(requestTracer, reader, writer);
				}
				resolve({
					reader: reader,
					writer: writer,
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
	private callbacks: Array<DataCallback> = [];

	constructor(socket: WebSocket) {
		super();
		this.socket = socket;

		socket.onmessage = (ev: MessageEvent) => {
			if (this.callbacks.length === 0) {
				this.fireError(new Error("message arrived on WebSocket but there is no listener"));
				return;
			}
			try {
				const data = JSON.parse(ev.data);
				this.callbacks.forEach(callback => callback(data));
			} catch (error) {
				this.fireError(error);
			}
		};
	}

	public listen(callback: DataCallback): void {
		this.callbacks.push(callback);
	}
}

/**
 * WebSocketMessageWriter wraps a WebSocket to conform to the MessageWriter interface.
 */
class WebSocketMessageWriter extends AbstractMessageWriter implements MessageWriter {
	private socket: WebSocket;
	private socketClosed: boolean;
	private errorCount: number;
	private callbacks: Array<DataCallback> = [];

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
		this.callbacks.forEach(callback => callback(msg));
		this.socket.send(JSON.stringify(msg));
	}

	public listen(callback: DataCallback): void {
		this.callbacks.push(callback);
	}
}

export interface MessageTrace {
	startTime: number;
	endTime: number;
	request: RequestMessage;
	response: ResponseMessage;
}

// vscode-jsonrpc2 doesn't export this interface, so we define it here.
// We also add meta which is a Sourcegraph specific extension.
interface ResponseMessage extends Message {
	id: number | string;
	result?: any;
	error?: any;
	meta?: { [key: string]: string; };
}

interface Listener {
	listen(callback: DataCallback): void;
}

function traceJSONRPCRequests(tracer: (trace: MessageTrace) => void, reader: Listener, writer: Listener): void {
	const inflight: Map<string | number, [RequestMessage, number]> = new Map();
	writer.listen((data: Message) => {
		const msg = data as RequestMessage;
		if (msg.id !== undefined) {
			inflight.set(msg.id, [msg, Date.now()]);
		}
	});
	reader.listen((data: Message) => {
		const resp = data as ResponseMessage;
		if (resp.id !== undefined) {
			const req = inflight.get(resp.id);
			if (req !== undefined) {
				inflight.delete(req[0].id);
				tracer({
					startTime: req[1],
					endTime: Date.now(),
					request: req[0],
					response: resp,
				});
			}
		}
	});
}

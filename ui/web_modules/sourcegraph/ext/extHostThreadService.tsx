import Event, { Emitter } from "vs/base/common/event";
import { TPromise } from "vs/base/common/winjs.base";
import { IMessagePassingProtocol } from "vs/base/parts/ipc/common/ipc";
import { IMainProcessExtHostIPC, create } from "vs/platform/extensions/common/ipcRemoteCom";
import { AbstractThreadService } from "vs/workbench/services/thread/common/abstractThreadService";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

declare var self: Worker;

/**
 * ExtHostThreadService is an implementation of IThreadService that
 * communicates with our MainThreadService using Worker.postMessage.
 */
export class ExtHostThreadService extends AbstractThreadService implements IThreadService {
	public _serviceBrand: any;
	private remoteCom: IMainProcessExtHostIPC;

	constructor() {
		super(false);

		const protocol = new WorkerProtocol();
		this.remoteCom = create(msg => protocol.send(msg));
		protocol.onMessage(msg => {
			this.remoteCom.handle(msg);
		});
		this.remoteCom.setManyHandler(this);
	}

	protected _callOnRemote(proxyId: string, path: string, args: any[]): TPromise<any> {
		return this.remoteCom.callOnRemote(proxyId, path, args);
	}
}

// WorkerProtocol communicates with a MainProtocol running in the main
// (non-Web Worker) context.
class WorkerProtocol implements IMessagePassingProtocol {
	private emitter: Emitter<any> = new Emitter<any>();
	get onMessage(): Event<any> { return this.emitter.event; }

	constructor() {
		if (self.onmessage) {
			throw new Error("worker.onmessage is already set");
		}
		self.onmessage = (message: MessageEvent) => {
			this.emitter.fire(message.data);
		};
		self.onerror = (err: ErrorEvent) => {
			console.error("Worker error:", err);
		};
	}

	send(message: any): void {
		self.postMessage(message);
	}
}

import Event, { Emitter } from "vs/base/common/event";
import * as strings from "vs/base/common/strings";
import { TPromise } from "vs/base/common/winjs.base";
import { IMessagePassingProtocol } from "vs/base/parts/ipc/common/ipc";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IMainProcessExtHostIPC, create } from "vs/platform/extensions/common/ipcRemoteCom";
import { AbstractThreadService } from "vs/workbench/services/thread/common/abstractThreadService";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

/**
 * MainThreadService is an implementation of IThreadService that
 * communicates using Worker.postMessage to a Web Worker extension
 * host.
 */
export class MainThreadService extends AbstractThreadService implements IThreadService {
	public _serviceBrand: any;

	private remoteCom: IMainProcessExtHostIPC;
	private logCommunication: boolean;
	private ready: boolean;

	constructor(
		@IEnvironmentService environmentService: IEnvironmentService,
	) {
		super(true);

		// Run `localStorage.logExtensionHostCommunication=true` in
		// your browser's JavaScript console to see detailed message
		// communication between window and extension host. Run
		// `delete localStorage.logExtensionHostCommunication` to
		// disable logging.
		this.logCommunication = environmentService.logExtensionHostCommunication || localStorage.getItem("logExtensionHostCommunication") !== null;
	}

	/**
	 * attachWorker sets the Web Worker that this MainThreadService
	 * communicates with. The protocol calls Worker.postMessage
	 * instead of window.postMessage so that messages go directly to
	 * the worker and are not broadcast to other potentially untrusted
	 * recipients.
	 */
	public attachWorker(worker: Worker): void {
		const protocol = new MainProtocol(worker);

		// Message: Window --> Extension Host
		this.remoteCom = create(msg => {
			if (this.logCommunication) {
				console.log("%c[Window \u2192 Extension]%c[len: " + strings.pad(msg.length, 5, " ") + "]", "color: darkgreen", "color: grey", msg); // tslint:disable-line no-console
			}

			protocol.send(msg);
		});

		// Message: Extension Host --> Window
		protocol.onMessage(msg => {
			if (this.logCommunication) {
				console.log("%c[Extension \u2192 Window]%c[len: " + strings.pad(msg.length, 5, " ") + "]", "color: darkgreen", "color: grey", msg); // tslint:disable-line no-console
			}

			this.remoteCom.handle(msg);
		});

		this.remoteCom.setManyHandler(this);

		this.ready = true;
	}

	protected _callOnRemote(proxyId: string, path: string, args: any[]): TPromise<any> {
		if (!this.ready) {
			throw new Error("protocol not ready (worker is not yet attached)");
		}
		return this.remoteCom.callOnRemote(proxyId, path, args);
	}
}

/**
 * MainProtocol communicates with a WorkerProtocol running in a Web
 * Worker context.
 */
class MainProtocol implements IMessagePassingProtocol {
	private emitter: Emitter<any> = new Emitter<any>();
	get onMessage(): Event<any> { return this.emitter.event; }

	constructor(private worker: Worker) {
		if (worker.onmessage) {
			throw new Error("worker.onmessage is already set");
		}
		worker.onmessage = (message: MessageEvent) => {
			this.emitter.fire(message.data);
		};
	}

	send(message: any): void {
		this.worker.postMessage(message);
	}
}

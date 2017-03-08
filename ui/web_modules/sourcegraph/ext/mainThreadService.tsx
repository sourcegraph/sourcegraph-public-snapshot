import Event, { Emitter } from "vs/base/common/event";
import * as strings from "vs/base/common/strings";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IMessagePassingProtocol } from "vs/base/parts/ipc/common/ipc";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IMainProcessExtHostIPC, create } from "vs/platform/extensions/common/ipcRemoteCom";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { AbstractThreadService } from "vs/workbench/services/thread/common/abstractThreadService";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

/**
 * MainThreadService is an implementation of IThreadService that
 * communicates using Worker.postMessage to a Web Worker extension
 * host.
 */
export class MainThreadService extends AbstractThreadService implements IThreadService {
	public _serviceBrand: any;

	private remotes: Map<string, IMainProcessExtHostIPC>;
	private logCommunication: boolean;

	constructor(
		@IEnvironmentService environmentService: IEnvironmentService,
		@IWorkspaceContextService private contextService: IWorkspaceContextService
	) {
		super(true);

		this.remotes = new Map<string, IMainProcessExtHostIPC>();

		// Run `localStorage.logExtensionHostCommunication=true` in
		// your browser's JavaScript console to see detailed message
		// communication between window and extension host. Run
		// `delete localStorage.logExtensionHostCommunication` to
		// disable logging.
		this.logCommunication = environmentService.logExtensionHostCommunication || localStorage.getItem("logExtensionHostCommunication") !== null;
	}

	/**
	 * attachWorker adds a Web Worker that this MainThreadService
	 * communicates with. The protocol calls Worker.postMessage
	 * instead of window.postMessage so that messages go directly to
	 * the worker and are not broadcast to other potentially untrusted
	 * recipients.
	 */
	public attachWorker(worker: Worker, workspace: URI): void {
		const protocol = new MainProtocol(worker);

		// Message: Window --> Extension Host
		const remoteCom = create(msg => {
			if (this.logCommunication) {
				console.log("%c[Window \u2192 " + workspace.path + "]%c[len: " + strings.pad(msg.length, 5, " ") + "]", "color: darkgreen", "color: grey", msg); // tslint:disable-line no-console
			}

			protocol.send(msg);
		});

		// Message: Extension Host --> Window
		protocol.onMessage(msg => {
			if (this.logCommunication) {
				console.log("%c[" + workspace.path + " \u2192 Window]%c[len: " + strings.pad(msg.length, 5, " ") + "]", "color: darkgreen", "color: grey", msg); // tslint:disable-line no-console
			}

			remoteCom.handle(msg);
		});

		remoteCom.setManyHandler(this);

		this.remotes.set(workspace.toString(), remoteCom);
		this.remotes.set(workspace.with({ query: `${workspace.query}~0` }).toString(), remoteCom);
	}

	protected _callOnRemote(proxyId: string, path: string, args: any[]): TPromise<any> {
		const routeToWorkspaceHost = uri => {
			const workspace = uri.with({ fragment: "" });
			const remoteCom = this.remotes.get(workspace.toString());
			if (!remoteCom) {
				throw new Error(`unable to route call ${proxyId}.${path} because no host for workspace ${workspace.toString()} (${this.remotes.size} hosts available)`);
			}
			return remoteCom.callOnRemote(proxyId, path, args);
		};
		switch (proxyId) {
			case "eExtHostLanguageFeatures":
				switch (path) {
					case "$provideReferences":
						return routeToWorkspaceHost(args[2] as URI);

					case "$provideWorkspaceReferences":
						return routeToWorkspaceHost(args[2] as URI);

					default:
						if (args.length >= 2 && args[1] instanceof URI) {
							return routeToWorkspaceHost(args[1] as URI);
						}
				}
				break;

			case "eExtHostDocuments":
				switch (path) {
					case "$provideTextDocumentContent":
						return routeToWorkspaceHost(args[1] as URI);

					case "$acceptModelAdd":
						return routeToWorkspaceHost(args[0].url as URI);
				}
				break;

			case "eExtHostEditors":
				switch (path) {
					case "$acceptTextEditorAdd":
						return routeToWorkspaceHost(args[0].document as URI);
				}
				break;
		}

		// Default to routing requests to the current workspace.
		return routeToWorkspaceHost(this.contextService.getWorkspace().resource);
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

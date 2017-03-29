import Event, { Emitter } from "vs/base/common/event";
import { IDisposable } from "vs/base/common/lifecycle";
import * as strings from "vs/base/common/strings";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IMessagePassingProtocol } from "vs/base/parts/ipc/common/ipc";
import { IEnvironmentService } from "vs/platform/environment/common/environment";
import { IRemoteCom, createProxyProtocol } from "vs/platform/extensions/common/ipcRemoteCom";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";
import { AbstractThreadService } from "vs/workbench/services/thread/common/abstractThreadService";
import { IThreadService } from "vs/workbench/services/thread/common/threadService";

import { getGitBaseResource, getWorkspaceForResource } from "sourcegraph/workbench/utils";

function asLoggingProtocol(protocol: IMessagePassingProtocol): IMessagePassingProtocol {

	protocol.onMessage(msg => {
		console.log("%c[Extension \u2192 Window]%c[len: " + strings.pad(msg.length, 5, " ") + "]", "color: darkgreen", "color: grey", msg); // tslint:disable-line no-console
	});

	return {
		onMessage: protocol.onMessage,

		send(msg: any): void {
			protocol.send(msg);
			console.log("%c[Window \u2192 Extension]%c[len: " + strings.pad(msg.length, 5, " ") + "]", "color: darkgreen", "color: grey", msg); // tslint:disable-line no-console
		}
	};
}

/**
 * MainThreadService is an implementation of IThreadService that
 * communicates using Worker.postMessage to a Web Worker extension
 * host.
 */
export class MainThreadService extends AbstractThreadService implements IThreadService {
	public _serviceBrand: any;

	private remotes: Map<string, IRemoteCom> = new Map<string, IRemoteCom>();
	private logCommunication: boolean;
	private workerEmitter: Emitter<string> = new Emitter<string>();

	constructor(
		@IEnvironmentService environmentService: IEnvironmentService,
		@IWorkspaceContextService private contextService: IWorkspaceContextService
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
	 * attachWorker adds a Web Worker that this MainThreadService
	 * communicates with. The protocol calls Worker.postMessage
	 * instead of window.postMessage so that messages go directly to
	 * the worker and are not broadcast to other potentially untrusted
	 * recipients.
	 */
	public attachWorker(worker: Worker, workspace: URI): void {
		let protocol: IMessagePassingProtocol = new MainProtocol(worker);
		if (this.logCommunication) {
			protocol = asLoggingProtocol(protocol);
		}

		const remoteCom = createProxyProtocol(protocol);

		remoteCom.setManyHandler(this);

		this.remotes.set(workspace.toString(), remoteCom);
		this.remotes.set(getGitBaseResource(workspace).toString(), remoteCom);
		this.workerEmitter.fire(workspace.toString());
		this.workerEmitter.fire(getGitBaseResource(workspace).toString());
	}

	protected _callOnRemote(proxyId: string, path: string, args: any[]): TPromise<any> {
		const routeToWorkspaceHostWithArgs = (uri: URI) => {
			const workspace = getWorkspaceForResource(uri);
			let remoteCom = this.remotes.get(workspace.toString());
			if (remoteCom) {
				return remoteCom.callOnRemote(proxyId, path, args);
			}
			// The main thread may need to wait a moment for a new workspace host to initialize.
			// We give the worker 3s to get set up before failing the request.
			return TPromise.wrap(new Promise((resolve, reject) => {
				let timedOut = false;
				let d: IDisposable;
				const timer = setTimeout(() => {
					timedOut = true;
					const matchingPaths: string[] = [];
					this.remotes.forEach((value, key) => {
						if (key.startsWith(workspace.toString())) {
							matchingPaths.push(key);
						}
					});
					if (d) {
						d.dispose();
					}
					reject(new Error(`unable to route call ${proxyId}.${path} because no host for workspace ${workspace.toString()} (${matchingPaths.length} hosts out of ${this.remotes.size} had matching paths: ${JSON.stringify(matchingPaths)})`));
				}, 3000);
				d = this.workerEmitter.event((e) => {
					remoteCom = this.remotes.get(workspace.toString());
					if (remoteCom && !timedOut) {
						clearTimeout(timer);
						if (d) {
							d.dispose();
						}
						resolve(remoteCom.callOnRemote(proxyId, path, args));
					}
				});
			}));
		};

		const routeToWorkspaceHost = (uri: URI) => routeToWorkspaceHostWithArgs(uri);
		const routeStringURLToWorkspaceHost = (stringURL: string) => routeToWorkspaceHostWithArgs(URI.parse(stringURL));

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

					default:
						return routeStringURLToWorkspaceHost(args[0]);
				}

			case "eExtHostEditors":
				switch (path) {
					case "$acceptTextEditorAdd":
						return routeToWorkspaceHost(args[0].document as URI);
					default:
						if (args[args.length - 1] instanceof URI) {
							return routeToWorkspaceHost(args[args.length - 1] as URI);
						}
				}
				break;

			case "eExtHostDocumentsAndEditors":
				switch (path) {
					case "$acceptDocumentsAndEditorsDelta":
						const arg = args[0];
						if (arg.addedDocuments && arg.addedDocuments.length > 0) {
							// assume the first doucment is for the same workspace as the rest
							return routeToWorkspaceHost(arg.addedDocuments[0].url as URI);
						} else if (arg.removedDocuments) {
							// assume again the first document is for the same workspace as the rest
							return routeStringURLToWorkspaceHost(arg.removedDocuments[0]);
						} else if (arg.addedEditors) {
							return routeToWorkspaceHost(arg.addedEditors[0].document as URI);
						}
						break;
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

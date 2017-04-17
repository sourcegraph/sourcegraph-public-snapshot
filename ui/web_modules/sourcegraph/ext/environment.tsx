import * as vscode from "vscode";

import URI from "vs/base/common/uri";
import { IWorkspaceRevState } from "vs/platform/workspace/common/workspace";

import { WorkRef, abbrevRef } from "libzap/lib/ref";
import { MessageStream } from "libzap/lib/remote/client";
import { URIUtils } from "sourcegraph/core/uri";
import { webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";
import { IEnvironment } from "vscode-zap/out/src/environment";

import { context } from "sourcegraph/app/context";

// VSCodeEnvironment is an implementation of IEnvironment used when
// running in the browser. It is backed by the vscode extension API
// and has access to browser APIs.
export class BrowserEnvironment implements IEnvironment {
	private docAtLastSave: Map<string, string> = new Map<string, string>();
	private docAtBase: Map<string, string> = new Map<string, string>(); // simulated doc at base commit (before ops)

	private toDispose: vscode.Disposable[] = [];

	constructor(private revState: IWorkspaceRevState) {
		// Track the initial contents of documents so we can revert.
		vscode.workspace.onDidOpenTextDocument((doc: vscode.TextDocument) => {
			if (doc.isDirty) {
				throw new Error(`expected to see document ${doc.uri.toString()} before it is dirty`);
			}
			this.updateDoc(this.docAtBase, doc);
			this.updateDoc(this.docAtLastSave, doc);
		}, null, this.toDispose);
		vscode.workspace.onDidSaveTextDocument(doc => this.onDidSaveTextDocument(doc), null, this.toDispose);
		vscode.workspace.onDidUpdateWorkspace(update => {
			// Save the last non-Zap rev, so that we can go back to a
			// sensible URL when Zap is turned off.
			if (update.revState) {
				this.revState = update.revState;
			}
		}, null, this.toDispose);
	}

	dispose(): void {
		this.toDispose.forEach(disposable => disposable.dispose());
		this.workRefResetEmitter.dispose();
	}

	private onDidSaveTextDocument(doc: vscode.TextDocument): void {
		this.updateDoc(this.docAtLastSave, doc);
	}

	private updateDoc(map: Map<string, string>, doc: vscode.TextDocument): void {
		const key = doc.uri.toString();
		map.set(key, doc.getText());
	}

	get rootURI(): vscode.Uri | undefined {
		return vscode.workspace.rootPath ? vscode.Uri.parse(vscode.workspace.rootPath) : undefined;
	}

	get repo(): string {
		return this.rootURI!.authority + this.rootURI!.path;
	}

	private workRefResetEmitter: vscode.EventEmitter<WorkRef | undefined> = new vscode.EventEmitter<WorkRef | undefined>();
	private _workRef: WorkRef | undefined;
	get onWorkRefReset(): vscode.Event<WorkRef | undefined> { return this.workRefResetEmitter.event; }
	public get workRef(): WorkRef | undefined { return this._workRef; }
	public setWorkRef(ref: WorkRef | undefined): void {
		const name = ref ? (ref.target || ref.name) : undefined;
		vscode.workspace.setWorkspaceState(this.rootURI!, {
			...this.revState,
			zapRef: name,
			zapRev: name ? abbrevRef(name) : undefined,
		});

		this._workRef = ref;
		this.workRefResetEmitter.fire(ref);
	}

	asRelativePathInsideWorkspace(uri: vscode.Uri): string | null {
		const workspace = this.rootURI!.toString();
		return URIUtils.tryConvertGitToFileURI(uri as URI).toString().substr(workspace.length + 1); // strip leading "/"
	}

	asAbsoluteURI(fileName: string): vscode.Uri {
		return URIUtils.createResourceURI(this.repo, fileName);
	}

	automaticallyApplyingFileSystemChanges: boolean = false;

	readTextDocumentOnDisk(uri: vscode.Uri): string {
		const text = this.docAtLastSave.get(uri.toString());
		if (text === undefined) {
			throw new Error(`no text for document at URI ${uri.toString()}`);
		}
		return text;
	}

	revertTextDocument(doc: vscode.TextDocument): Thenable<any> {
		// In the web context, documents are reverted via the vscode text document API.
		// See Controller.revertAllDocuemnts.
		//
		// IN the vscode context, there are other codepaths to reset documents that
		// require this method to have a non-empty implementation.
		return Promise.resolve();
	}

	revertAllTextDocuments(): Thenable<any> {
		return new Promise(async (resolve) => {
			const docsToRevert = new Set<string>(vscode.workspace.textDocuments.map(doc => {
				// Only revert mutable docs (not `git://github.com/gorilla/mux?sha#file/path` documents).
				// Only revert docs if a .revert() method exists on the vscode text document
				if ((doc as any).revert && doc.uri.query === "") {
					return doc.uri.toString();
				}
				return "";
			}).filter(doc => doc !== ""));

			// The zap extension has a race condition for reverting documents and applying
			// new ops. If a user switches from zap ref A to B, we must revert all documents
			// before applying ops for B. Therefore, the zap extension must wait to be notified
			// from the main thread that documents have actually been reverted.
			let disposable: vscode.Disposable | undefined;
			if ((vscode.workspace as any).onDidRevertTextDocument) {
				disposable = (vscode.workspace as any).onDidRevertTextDocument((docUri: string) => {
					docsToRevert.delete(docUri);
					if (docsToRevert.size === 0) {
						if (disposable) {
							disposable.dispose();
						}
						resolve(); // all documents have been reverted, this function has completed
					}
				});
			}

			await Promise.all(vscode.workspace.textDocuments
				.filter(doc => docsToRevert.has(doc.uri.toString()))
				.map(doc => (doc as any).revert()));
		});
	}

	deleteTextDocument(doc: vscode.TextDocument): Thenable<void> {
		return doc.delete();
	}

	openChannel(id: string): Thenable<MessageStream> {
		if (id !== "zap") {
			throw new Error(`unknown channel id: ${id}`);
		}

		// self.location is the blob: URI, so we need to get the main page location.
		let wsOrigin = self.location.origin.replace(/^https?:\/\//, (match) => {
			return match === "http://" ? "ws://" : "wss://";
		});
		if (wsOrigin === "wss://sourcegraph.com") {
			wsOrigin = "wss://ws.sourcegraph.com";
		}
		return webSocketStreamOpener(`${wsOrigin}/.api/zap`)();
	}

	get userID(): string {
		const ctx: typeof context = self["sourcegraphContext"];
		const user = ctx && ctx.user ? ctx.user.Login : "anonymous";
		return `${user}@web`;
	}
}

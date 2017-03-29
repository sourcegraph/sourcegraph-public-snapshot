import * as vscode from "vscode";

import URI from "vs/base/common/uri";

import { MessageStream } from "libzap/lib/remote/client";
import { URIUtils } from "sourcegraph/core/uri";
import { webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";
import { IEnvironment } from "vscode-zap/out/src/environment";

import { context } from "sourcegraph/app/context";

// VSCodeEnvironment is an implementation of IEnvironment used when
// running in the browser. It is backed by the vscode extension API
// and has access to browser APIs.
class BrowserEnvironment implements IEnvironment {
	private docAtLastSave: Map<string, string> = new Map<string, string>();
	private docAtBase: Map<string, string> = new Map<string, string>(); // simulated doc at base commit (before ops)
	private _prevZapRef: string | undefined;
	private _zapRef: string | undefined;
	private _isRunning: boolean;

	constructor() {
		// Track the initial contents of documents so we can revert.
		vscode.workspace.onDidOpenTextDocument((doc: vscode.TextDocument) => {
			if (doc.isDirty) {
				throw new Error(`expected to see document ${doc.uri.toString()} before it is dirty`);
			}
			this.updateDoc(this.docAtBase, doc);
			this.updateDoc(this.docAtLastSave, doc);
		});
		vscode.workspace.onDidSaveTextDocument(doc => this.onDidSaveTextDocument(doc));
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

	private zapRefChangeEmitter: vscode.EventEmitter<string | undefined> = new vscode.EventEmitter<string | undefined>();
	get onDidChangeZapRef(): vscode.Event<string | undefined> { return this.zapRefChangeEmitter.event; }

	get zapRef(): string | undefined {
		return this._zapRef;
	}

	set zapRef(ref: string | undefined) {
		this._prevZapRef = this._zapRef;
		this._zapRef = ref;
		this.zapRefChangeEmitter.fire(ref);
		this.zapBranchChangeEmitter.fire(ref ? ref.replace(/^branch\//, "") : ref);
	}

	get prevZapRef(): string | undefined {
		return this._prevZapRef;
	}

	set prevZapRef(ref: string | undefined) {
		this._prevZapRef = ref;
	}

	// On the web, the Zap branch is ALWAYS the Zap ref stripped of the "branch/" prefix.
	get zapBranch(): string | undefined {
		return this._zapRef ? this._zapRef.replace(/^branch\//, "") : this._zapRef;
	}
	private zapBranchChangeEmitter: vscode.EventEmitter<string | undefined> = new vscode.EventEmitter<string | undefined>();
	get onDidChangeZapBranch(): vscode.Event<string | undefined> {
		return this.zapBranchChangeEmitter.event;
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

	revertTextDocumentToBase(doc: vscode.TextDocument): Thenable<any> {
		// In the web context, documents are reverted via the vscode text document API.
		// See Controller.revertAllDocuemnts.
		//
		// IN the vscode context, there are other codepaths to reset documents that
		// require this method to have a non-empty implementation.
		return Promise.resolve();
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
		return webSocketStreamOpener(`${wsOrigin}/.api/zap`)();
	}

	get userID(): string {
		const ctx: typeof context = self["sourcegraphContext"];
		const user = ctx && ctx.user ? ctx.user.Login : "anonymous";
		return `${user}@web`;
	}

	set isRunning(status: boolean) {
		this._isRunning = status;
	}

	get isRunning(): boolean {
		return this._isRunning;
	}
}

export default new BrowserEnvironment(); // tslint:disable-line no-default-export

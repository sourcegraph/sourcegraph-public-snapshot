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
	private _zapRef: string;
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

		this.zapRef = self["__tmpZapRef"];
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
		return URIUtils.repoParams(this.rootURI as URI).repo;
	}

	private zapRefChangeEmitter: vscode.EventEmitter<string> = new vscode.EventEmitter<string>();
	get onDidChangeZapRef(): vscode.Event<string> { return this.zapRefChangeEmitter.event; }

	get zapRef(): string {
		return this._zapRef;
	}

	set zapRef(ref: string) {
		this._zapRef = ref;
		this.zapRefChangeEmitter.fire(ref);
		vscode.commands.executeCommand("zap.reference.change", ref);
	}

	get zapBranch(): string { return this._zapRef; }
	get onDidChangeZapBranch(): vscode.Event<string> { return this.zapRefChangeEmitter.event; }

	asRelativePathInsideWorkspace(uri: vscode.Uri): string | null {
		if (uri.scheme !== "git") { return null; }
		// TODO(sqs): Check that uri is underneath the rootURI.
		return URIUtils.repoParams(uri as URI).path;
	}

	asAbsoluteURI(fileName: string): vscode.Uri {
		const { repo, rev } = URIUtils.repoParams(this.rootURI! as URI);
		return URIUtils.pathInRepo(repo, rev, fileName);
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
		return this.doRevertTextDocument(this.docAtLastSave, doc);
	}

	revertTextDocumentToBase(doc: vscode.TextDocument): Thenable<any> {
		return this.doRevertTextDocument(this.docAtBase, doc);
	}

	private doRevertTextDocument(contents: Map<string, string>, doc: vscode.TextDocument): Thenable<any> {
		const initialContents = contents.get(doc.uri.toString());
		if (initialContents === undefined) {
			throw new Error(`revertTextDocument: unknown initial contents for ${doc.uri.toString()}`);
		}

		const edit = new vscode.WorkspaceEdit();
		const entireRange = new vscode.Range(new vscode.Position(0, 0), doc.positionAt(doc.getText().length));
		edit.replace(doc.uri, entireRange, initialContents);
		return vscode.workspace.applyEdit(edit).then(() => doc.save());
	}

	openChannel(id: string): Thenable<MessageStream> {
		if (id !== "zap") {
			throw new Error(`unknown channel id: ${id}`);
		}

		const ctx: typeof context = self["sourcegraphContext"];
		return webSocketStreamOpener(`${ctx.wsURL}/.api/zap`)();
	}

	get userID(): string {
		const ctx: typeof context = self["sourcegraphContext"];
		const user = ctx && ctx.user ? ctx.user.Login : "anonymous";
		return `${user}@web`;
	}

	set isRunning(status: boolean) {
		this._isRunning = status;
		vscode.commands.executeCommand("zap.status.change", status);
	}

	get isRunning(): boolean {
		return this._isRunning;
	}
}

export default new BrowserEnvironment(); // tslint:disable-line no-default-export

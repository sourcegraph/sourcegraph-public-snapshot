import * as vscode from "vscode";

import URI from "vs/base/common/uri";

import { MessageStream } from "libzap/lib/remote/client";
import { URIUtils } from "sourcegraph/core/uri";
import { webSocketStreamOpener } from "sourcegraph/ext/lsp/connection";
import { IEnvironment } from "vscode-zap/out/src/environment";

// VSCodeEnvironment is an implementation of IEnvironment used when
// running in the browser. It is backed by the vscode extension API
// and has access to browser APIs.
class BrowserEnvironment implements IEnvironment {
	private docAtLastSave: Map<string, string> = new Map<string, string>();
	private _zapRef: string;

	constructor() {
		// Track the initial contents of documents so we can revert.
		vscode.workspace.onDidOpenTextDocument((doc: vscode.TextDocument) => {
			this.updateDocAtLastSave(doc);
		});
		vscode.workspace.onDidSaveTextDocument((doc: vscode.TextDocument) => {
			this.updateDocAtLastSave(doc);
		});

		this.zapRef = self["__tmpZapRef"];
	}

	private updateDocAtLastSave(doc: vscode.TextDocument): void {
		const key = doc.uri.toString();
		if (this.docAtLastSave.has(key)) {
			throw new Error(`expected to see document ${key} only once`);
		}
		if (doc.isDirty) {
			throw new Error(`expected to see document ${key} before it is dirty`);
		}
		this.docAtLastSave.set(key, doc.getText());
	}

	get rootURI(): vscode.Uri | undefined {
		return vscode.workspace.rootPath ? vscode.Uri.parse(vscode.workspace.rootPath) : undefined;
	}

	get repo(): string {
		return URIUtils.repoParams(this.rootURI as URI).repo;
	}

	private zapRefChangeEmitter: vscode.EventEmitter<string> = new vscode.EventEmitter<string>();
	get onDidChangeZapRef(): vscode.Event<string> { return this.zapRefChangeEmitter.event; } // never fires TODO(sqs)
	// get zapRef(): string {
	// 	// TODO(sqs): hackily get the passed-through zap ref
	// 	return self["__tmpZapRef"] || "master@sqs";

	// 	// // TODO(sqs): this will get the absolute commit ID, not the branch
	// 	// const rev = URIUtils.repoParams(this.rootURI as URI).rev;
	// 	// if (!rev) {
	// 	// 	throw new Error(`no rev in rootURI: ${this.rootURI!.toString()}`);
	// 	// }
	// 	// return rev;
	// }

	get zapRef(): string {
		return this._zapRef;
	}

	set zapRef(ref: string) {
		this._zapRef = ref;
		this.zapRefChangeEmitter.fire(ref);
	}

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

	revertTextDocument2(doc: vscode.TextDocument): Thenable<any> {
		// HACK(sqs): see the comment in the lone call site of
		// revertTextDocument2 in zap. this does not happen in the
		// web, so we can noop here.
		return Promise.resolve(null);
	}

	revertTextDocument(doc: vscode.TextDocument): Thenable<any> {
		const initialContents = this.docAtLastSave.get(doc.uri.toString());
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

		// self.location is the blob: URI, so we need to get the main page location.
		let wsOrigin = self.location.origin.replace(/^https?:\/\//, (match) => {
			return match === "http://" ? "ws://" : "wss://";
		});
		return webSocketStreamOpener(`${wsOrigin}/.api/zap`)();
	}
}

export default new BrowserEnvironment(); // tslint:disable-line no-default-export

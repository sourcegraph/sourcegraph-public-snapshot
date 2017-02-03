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
	get rootURI(): vscode.Uri | undefined {
		return vscode.workspace.rootPath ? vscode.Uri.parse(vscode.workspace.rootPath) : undefined;
	}

	get repo(): string {
		return URIUtils.repoParams(this.rootURI as URI).repo;
	}

	get rev(): string {
		// TODO(sqs): hackily get the passed-through zap ref
		return self["__tmpZapRef"] || "master@sqs";

		// // TODO(sqs): this will get the absolute commit ID, not the branch
		// const rev = URIUtils.repoParams(this.rootURI as URI).rev;
		// if (!rev) {
		// 	throw new Error(`no rev in rootURI: ${this.rootURI!.toString()}`);
		// }
		// return rev;
	}

	asRelativePathInsideWorkspace(uri: vscode.Uri): string | null {
		// TODO(sqs): Check that uri is underneath the rootURI.
		return URIUtils.repoParams(uri as URI).path;
	}

	asAbsoluteURI(fileName: string): vscode.Uri {
		const { repo, rev } = URIUtils.repoParams(this.rootURI! as URI);
		return URIUtils.pathInRepo(repo, rev, fileName);
	}

	textDocumentIsDirtyHack(doc: vscode.TextDocument): boolean {
		// TODO(sqs): this doc.version>0 is a hack and is not correct
		// in general - it can be version 5 but it was just
		// saved. need to track dirty for real in the web app.
		return doc.isDirty || doc.version > 0;
	}

	automaticallyApplyingFileSystemChanges: boolean = false;

	revertTextDocument2(doc: vscode.TextDocument): Thenable<any> {
		console.warn("revertTextDocument2 is not yet implemented in the browser"); // tslint:disable-line no-console
		return Promise.resolve(null);
	}

	revertTextDocument(doc: vscode.TextDocument): Thenable<any> {
		console.warn("revertTextDocument is not yet implemented in the browser"); // tslint:disable-line no-console
		return Promise.resolve(null);
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

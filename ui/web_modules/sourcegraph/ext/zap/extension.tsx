import * as vscode from "vscode";

import browserEnvironment from "sourcegraph/ext/environment";
import { Controller } from "vscode-zap/out/src/controller";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

let controller: Controller | null;

export function activate(): any {
	const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
	controller = activateCommon(browserEnvironment, ctx);
	if (controller) {
		// TODO/HACK: Pass information through file service instead of arbitrary commands. @Kingy
		vscode.commands.registerCommand("sourcegraph.resolve.file", (resource: vscode.Uri) => {
			let document = vscode.workspace.textDocuments
				.filter(doc => doc.uri.toString() === resource.toString());
			if (document.length === 1) {
				let doc = document[0];
				if (doc.uri.toString() === resource.toString() && controller!.environment.asRelativePathInsideWorkspace(resource) !== null) {
					return doc.getText();
				}
			}

			return null;
		});

		vscode.commands.executeCommand("sourcegraph.extensions.loaded");
	}
	return controller;
}

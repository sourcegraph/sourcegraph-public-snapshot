import * as vscode from "vscode";

import browserEnvironment from "sourcegraph/ext/environment";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

export function activate(): any {
	if (!browserEnvironment.rev) {
		console.debug("Zap disabled because there is no ?tmpZapRef= URL query parameter."); // tslint:disable-line no-console
		return;
	}
	const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
	return activateCommon(browserEnvironment, ctx);
}

import * as vscode from "vscode";

import browserEnvironment from "sourcegraph/ext/environment";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

export function activate(): any {
	const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
	return activateCommon(browserEnvironment, ctx);
}

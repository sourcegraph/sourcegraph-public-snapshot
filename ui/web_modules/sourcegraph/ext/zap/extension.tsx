import { isOnPremInstance } from "sourcegraph/app/context";
import * as vscode from "vscode";

import browserEnvironment from "sourcegraph/ext/environment";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

import { InitializationOptions } from "sourcegraph/ext/protocol";

export function activate(): any {
	const initOpts: InitializationOptions = (self as any).extensionHostOptions;
	if (isOnPremInstance(initOpts.context.authEnabled)) {
		return;
	}
	const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
	return activateCommon(browserEnvironment, ctx, initOpts, true);
}

import * as vscode from "vscode";

import browserEnvironment from "sourcegraph/ext/environment";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

import { InitializationOptions } from "sourcegraph/ext/protocol";

export function activate(): any {
	const initOpts: InitializationOptions = (self as any).extensionHostOptions;
	if (initOpts.context.appURL === "http://sourcegraph.dev.uberinternal.com:30000" || initOpts.context.appURL === "http://node.aws.sgdev.org:30000") {
		return;
	}
	const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
	return activateCommon(browserEnvironment, ctx, initOpts.revState);
}

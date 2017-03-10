import * as vscode from "vscode";

import { context } from "sourcegraph/app/context";
import browserEnvironment from "sourcegraph/ext/environment";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

import { InitializationOptions } from "sourcegraph/ext/protocol";

export function activate(): any {
	if (context.appURL === "http://sourcegraph.dev.uberinternal.com:30000" || context.appURL === "http://node.aws.sgdev.org:30000") {
		return;
	}
	const initOpts: InitializationOptions = (self as any).extensionHostOptions;
	const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
	return activateCommon(browserEnvironment, ctx, initOpts.revState);
}

import { isOnPremInstance } from "sourcegraph/app/context";
import * as vscode from "vscode";

import { BrowserEnvironment } from "sourcegraph/ext/environment";
import { activate as activateCommon } from "vscode-zap/out/src/extension.common";

import { NewRef, Ref } from "libzap/lib/ref";
import { InitializationOptions } from "sourcegraph/ext/protocol";
import { Features } from "sourcegraph/util/features";

export function activate(): void {
	if (!Features.zap.isEnabled()) {
		return;
	}
	const initOpts: InitializationOptions = (self as any).extensionHostOptions;
	if (isOnPremInstance(initOpts.context.authEnabled)) {
		return;
	}

	if (initOpts && initOpts.revState) {
		const ctx: vscode.ExtensionContext = { subscriptions: [] as vscode.Disposable[] } as any;
		const env = new BrowserEnvironment(initOpts.revState);

		// Synthesize initial work ref.
		const workRef: Ref | NewRef = ({ name: "head/local" } as any);
		if (initOpts.revState.zapRef) {
			workRef.state = { target: initOpts.revState.zapRef };
		} else if (initOpts.revState.commitID && initOpts.revState.branch) {
			workRef.state = {
				data: {
					gitBase: initOpts.revState.commitID,
					gitBranch: initOpts.revState.branch,
					history: [],
				},
			};
		} else {
			// Unable to initialize Zap because there is not enough
			// information about the current resource.
		}

		activateCommon(env, ctx, workRef, initOpts, true);
	}
}

/* --------------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See License.txt in the project root for license information.
 * ------------------------------------------------------------------------------------------ */
'use strict';

import * as path from 'path';

import { workspace, Disposable, ExtensionContext } from 'vscode';
import { LanguageClient, LanguageClientOptions, SettingMonitor, ServerOptions, ErrorAction, CloseAction, TransportKind } from 'vscode-languageclient';

export function activate(context: ExtensionContext) {
	let serverOptions: ServerOptions = {
		run : { command: "sample_server" },
		debug : { command: "sample_server" },
	}

	// Options to control the language client
	let clientOptions: LanguageClientOptions = {
		// Register the server for plain text documents
		documentSelector: ['plaintext'],
		synchronize: {
			// Synchronize the setting section 'languageServerExample' to the server
			configurationSection: 'languageServerExample',
		},
		errorHandler: {
			error(error: Error, message, count: number): ErrorAction {
				console.log("ERR", error, message, count);
				return ErrorAction.Shutdown;
			},
    		closed(): CloseAction {
				return CloseAction.Restart;
			},
		},
	}

	// Create the language client and start the client.
	let disposable = new LanguageClient('Language Server Example', serverOptions, clientOptions).start();

	// Push the disposable to the context's subscriptions so that the
	// client can be deactivated on extension deactivation
	context.subscriptions.push(disposable);
}

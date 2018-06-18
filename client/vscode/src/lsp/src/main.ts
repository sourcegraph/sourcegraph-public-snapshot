/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Sourcegraph. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
'use strict'

import * as vscode from 'vscode'
import { newClient } from './client'
import * as log from './log'
import * as child_process from 'child_process'

const runGitCommand = command =>
    child_process.execSync(command, { cwd: vscode.workspace.workspaceFolders[0].uri.fsPath }).toString()

const getRemoteUrl = (): vscode.Uri => {
    const result = runGitCommand('git remote --verbose')
    const regex = /^([^\s]+)\s+([^\s]+)\s/
    const rawRemotes = result
        .trim()
        .split('\n')
        .filter(b => !!b)
        .map(line => regex.exec(line))
        .filter(g => !!g)
        .map((groups: RegExpExecArray) => ({ name: groups[1], url: groups[2] }))

    const matches = /^git@([^:]*):(.*)$/.exec(rawRemotes[0].url)

    return vscode.Uri.parse(`https://${matches[1]}/${matches[2]}`)
}

const getGitCommit = (): string => {
    return runGitCommand('git rev-parse HEAD').trim()
}

export async function activate(context: vscode.ExtensionContext): Promise<void> {
    // TODO(chris) register language clients based on extensions in the Sourcegraph registry

    // Language IDs (e.g. shellscript here) come from https://code.visualstudio.com/docs/languages/identifiers
    const client = newClient('bash', ['shellscript'], getRemoteUrl(), getGitCommit())

    client.start()

    client.onReady().then(
        () => {
            log.outputChannel.appendLine('LanguageClient ready.')
        },
        err => {
            log.outputChannel.appendLine(`Error activating LSP root: ${err}.`)
        }
    )
}

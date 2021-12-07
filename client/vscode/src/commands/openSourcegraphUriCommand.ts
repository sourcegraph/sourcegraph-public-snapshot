import * as vscode from 'vscode'

import { SourcegraphFileSystemProvider } from '../file-system/SourcegraphFileSystemProvider'
import { CompareRange, SourcegraphUri } from '../file-system/SourcegraphUri'
import { log } from '../log'

export async function openSourcegraphUriCommand(fs: SourcegraphFileSystemProvider, uri: SourcegraphUri): Promise<void> {
    if (uri.compareRange) {
        await openCompareUri(uri, uri.compareRange)
        return
    }
    if (!uri.revision) {
        const metadata = await fs.repositoryMetadata(uri.repositoryName)
        uri = uri.withRevision(metadata?.defaultBranch || 'HEAD')
    }
    const textDocument = await vscode.workspace.openTextDocument(vscode.Uri.parse(uri.uri))
    const selection = getSelection(uri, textDocument)
    await vscode.window.showTextDocument(textDocument, {
        selection,
        viewColumn: vscode.ViewColumn.Active,
    })
}

async function openCompareUri(uri: SourcegraphUri, compareRange: CompareRange): Promise<void> {
    try {
        // TODO: fix me.
        log.appendLine(`openCompareUri uri=${uri.uri} compareRange=${JSON.stringify(compareRange)}`)
        await vscode.commands.executeCommand(
            'vscode.diff',
            vscode.Uri.parse('sourcegraph://sourcegraph.com/github.com/scalameta/metals@v0.10.0/-/blob/build.sbt'),
            vscode.Uri.parse('sourcegraph://sourcegraph.com/github.com/scalameta/metals@v0.10.7/-/blob/build.sbt'),
            'build.sbt (v0.10.0 <-> v0.10.7)'
        )
    } catch (error) {
        log.error(`openCompareUri(${uri.uri})`, error)
    }
}

function getSelection(uri: SourcegraphUri, textDocument: vscode.TextDocument): vscode.Range | undefined {
    if (typeof uri?.position?.line !== 'undefined' && typeof uri?.position?.character !== 'undefined') {
        return offsetRange(uri.position.line - 1, uri.position.character)
    }
    if (typeof uri?.position?.line !== 'undefined') {
        return offsetRange(uri.position.line - 1, 0)
    }

    // There's no explicitly provided line number. Instead of focusing on the
    // first line (which usually contains lots of imports), we use a heuristic
    // to guess the location where the "main symbol" is defined (a
    // function/class/struct/interface with the same name as the filename).
    if (uri.path && isFilenameThatMayDefineSymbols(uri.path)) {
        const fileNames = uri.path.split('/')
        const fileName = fileNames[fileNames.length - 1]
        const symbolName = fileName.split('.')[0]
        const text = textDocument.getText()
        const symbolMatches = new RegExp(` ${symbolName}\\b`).exec(text)
        if (symbolMatches) {
            const position = textDocument.positionAt(symbolMatches.index + 1)
            return new vscode.Range(position, position)
        }
    }

    return undefined
}

function offsetRange(line: number, character: number): vscode.Range {
    const position = new vscode.Position(line, character)
    return new vscode.Range(position, position)
}

/**
 * @returns true if this file may contain code from a programming language that
 * defines symbol.
 */
function isFilenameThatMayDefineSymbols(path: string): boolean {
    return !(path.endsWith('.md') || path.endsWith('.markdown') || path.endsWith('.txt') || path.endsWith('.log'))
}

/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { FormattingOptions, Segment } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import path from 'path'
import * as sourcegraph from 'sourcegraph'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { ExecServerClient } from '../execServer/client'
import { ResolvedDependencyInPackage } from './packageManager'

const MINIMAL_WORKTREE = true

export const editForDependencyUpgrade = async (
    {
        packageJson,
        lockfile,
        dependency,
    }: Pick<ResolvedDependencyInPackage, 'dependency'> & {
        packageJson: Pick<sourcegraph.TextDocument, 'uri' | 'text'>
        lockfile: Pick<sourcegraph.TextDocument, 'uri' | 'text'>
    },
    commands: string[][],
    execServerClient: ExecServerClient
): Promise<sourcegraph.WorkspaceEdit> => {
    const p = parseRepoURI(packageJson.uri)
    const packageJsonName = path.basename(parseRepoURI(packageJson.uri).filePath!)
    const lockfileName = path.basename(parseRepoURI(lockfile.uri).filePath!)
    const result = await execServerClient({
        commands,
        dir: path.dirname(p.filePath!),
        ...(MINIMAL_WORKTREE
            ? {
                  files: {
                      [packageJsonName]: packageJson.text!,
                      [lockfileName]: lockfile.text!,
                  },
              }
            : {
                  context: {
                      repository: p.repoName,
                      commit: p.commitID!,
                  },
              }),
    })

    for (const command of result.commands) {
        if (!command.ok) {
            throw new Error(
                `error upgrading dependency '${dependency.name}' in ${packageJson.uri}: ${command.error}\n${command.combinedOutput}`
            )
        }
    }

    if (MINIMAL_WORKTREE) {
        const edit = new sourcegraph.WorkspaceEdit()
        edit.set(new URL(packageJson.uri), [sourcegraph.TextEdit.patch(result.fileDiffs![packageJsonName])])
        edit.set(new URL(lockfile.uri), [sourcegraph.TextEdit.patch(result.fileDiffs![lockfileName])])
        return edit
    }

    throw new Error('only MINIMAL_WORKTREE is supported')
    // return computeDiffs([
    //     { old: packageJson, newText: result.files![packageJsonName] },
    //     { old: lockfile, newText: result.files![lockfileName] },
    // ])
}

function computeDiffs(files: { old: sourcegraph.TextDocument; newText?: string }[]): sourcegraph.WorkspaceEdit {
    const edit = new sourcegraph.WorkspaceEdit()
    for (const { old, newText } of files) {
        // TODO!(sqs): handle creation/removal
        if (old.text !== undefined && newText !== undefined && old.text !== newText) {
            edit.replace(
                new URL(old.uri),
                new sourcegraph.Range(new sourcegraph.Position(0, 0), old.positionAt(old.text.length)),
                newText
            )
        }
    }
    return edit
}

const PACKAGE_JSON_FORMATTING_OPTIONS: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

export const editPackageJson = (
    doc: sourcegraph.TextDocument,
    operations: { path: Segment[]; value: any }[],
    workspaceEdit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit => {
    for (const op of operations) {
        const edits = setProperty(doc.text!, op.path, op.value, PACKAGE_JSON_FORMATTING_OPTIONS)
        for (const edit of edits) {
            workspaceEdit.replace(
                new URL(doc.uri),
                new sourcegraph.Range(doc.positionAt(edit.offset), doc.positionAt(edit.offset + edit.length)),
                edit.content
            )
        }
    }
    return workspaceEdit
}

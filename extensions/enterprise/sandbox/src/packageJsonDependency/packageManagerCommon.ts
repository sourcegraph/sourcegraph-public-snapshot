/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { PackageJsonPackage, PackageJsonDependency } from './packageManager'
import path from 'path'
import * as sourcegraph from 'sourcegraph'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { ExecServerClient } from '../execServer/client'

const MINIMAL_WORKTREE = true

export const editForDependencyUpgrade = async (
    pkg: PackageJsonPackage,
    dep: PackageJsonDependency,
    commands: string[][],
    execServerClient: ExecServerClient
): Promise<sourcegraph.WorkspaceEdit> => {
    const p = parseRepoURI(pkg.packageJson.uri)
    const packageJsonName = path.basename(parseRepoURI(pkg.packageJson.uri).filePath!)
    const lockfileName = path.basename(parseRepoURI(pkg.lockfile.uri).filePath!)
    const result = await execServerClient({
        commands,
        dir: path.dirname(p.filePath!),
        ...(MINIMAL_WORKTREE
            ? {
                  files: {
                      [packageJsonName]: pkg.packageJson.text!,
                      [lockfileName]: pkg.lockfile.text!,
                  },
              }
            : {
                  context: {
                      repository: p.repoName,
                      commit: p.commitID!,
                  },
              }),
    })
    return computeDiffs([
        { old: pkg.packageJson, newText: result.files[packageJsonName] },
        { old: pkg.lockfile, newText: result.files[lockfileName] },
    ])
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

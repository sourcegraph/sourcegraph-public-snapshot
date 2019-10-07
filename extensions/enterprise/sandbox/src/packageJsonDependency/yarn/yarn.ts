/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { PackageJsonPackageManager, PackageJsonPackage } from '../packageManager'
import path from 'path'
import semver from 'semver'
import { flatten } from 'lodash'
import { memoizedFindTextInFiles } from '../../util'
import { from } from 'rxjs'
import { toArray } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../../shared/src/util/types'
import { parseRepoURI } from '../../../../../../shared/src/util/url'
import { createExecServerClient } from '../../execServer/client'
import { yarnLogicalTree } from './logicalTree'

const yarnExecClient = createExecServerClient('a8n-yarn-exec', ['package.json', 'yarn.lock'])

export const yarnPackageManager: PackageJsonPackageManager = {
    packagesWithUnsatisfiedDependencyVersionRange: async ({ name, version }) => {
        const parsedVersionRange = new semver.Range(version)

        const results = flatten(
            await from(
                memoizedFindTextInFiles(
                    {
                        pattern: `\\b${name}\\b`,
                        type: 'regexp',
                    },
                    {
                        repositories: {
                            includes: [],
                            type: 'regexp',
                        },
                        files: {
                            includes: ['(^|/)yarn.lock$'],
                            excludes: ['node_modules'],
                            type: 'regexp',
                        },
                        maxResults: 100, // TODO!(sqs): increase
                    }
                )
            )
                .pipe(toArray())
                .toPromise()
        )

        const check = async (result: sourcegraph.TextSearchResult): Promise<PackageJsonPackage | null> => {
            const packageJson = await sourcegraph.workspace.openTextDocument(
                new URL(result.uri.replace(/yarn\.lock$/, 'package.json'))
            )
            const lockfile = await sourcegraph.workspace.openTextDocument(new URL(result.uri))
            try {
                const dep = getYarnLockDependency(packageJson.text!, lockfile.text!, name)
                if (!dep) {
                    return null
                }
                return semver.satisfies(dep.version, parsedVersionRange) ? null : { packageJson, lockfile }
            } catch (err) {
                console.error(`Error checking yarn.lock and package.json for ${result.uri}.`, err, {
                    lockfile: lockfile.text,
                    packagejson: packageJson.text,
                })
                return null
            }
        }
        return (await Promise.all(results.map(check))).filter(isDefined)
    },

    editForDependencyUpgrade: async (pkg, dep) => {
        const p = parseRepoURI(pkg.packageJson.uri)
        const result = await yarnExecClient({
            commands: [['yarn', 'upgrade', '--', `${dep.name}@${dep.version}`]],
            context: {
                repository: p.repoName,
                commit: p.commitID!,
                path: path.dirname(p.filePath!),
            },
            // TODO!(sqs): dir
        })
        return computeDiffs([
            { old: pkg.packageJson, newText: result.files['package.json'] },
            { old: pkg.lockfile, newText: result.files['yarn.lock'] },
        ])
    },
}

function computeDiffs(files: { old: sourcegraph.TextDocument; newText?: string }[]): sourcegraph.WorkspaceEdit {
    const edit = new sourcegraph.WorkspaceEdit()
    for (const { old, newText } of files) {
        // TODO!(sqs): handle creation/removal
        if (old.text !== undefined && newText !== undefined && old.text !== newText) {
            edit.replace(
                new URL(old.uri),
                new sourcegraph.Range(new sourcegraph.Position(0, 0), old.positionAt(old.text!.length)),
                newText
            )
        }
    }
    return edit
}

function getYarnLockDependency(packageJson: string, yarnLock: string, packageName: string): { version: string } | null {
    const tree = yarnLogicalTree(JSON.parse(packageJson), yarnLock)
    let found: any
    tree.forEach((dep, next) => {
        if (dep.name === packageName) {
            found = dep
        } else {
            next()
        }
    })
    return found ? { version: found.version } : null
}

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
import { lockTree } from './logicalTree'

const npmExecClient = createExecServerClient('a8n-npm-exec', ['package.json', 'package-lock.json'])

export const npmPackageManager: PackageJsonPackageManager = {
    packagesWithUnsatisfiedDependencyVersionRange: async ({ name, version }) => {
        const parsedVersionRange = new semver.Range(version)

        const results = flatten(
            await from(
                memoizedFindTextInFiles(
                    {
                        pattern: `"${name}"`,
                        type: 'regexp',
                    },
                    {
                        repositories: {
                            includes: [],
                            type: 'regexp',
                        },
                        files: {
                            includes: ['(^|/)package-lock.json$'],
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
                new URL(result.uri.replace(/package-lock\.json$/, 'package.json'))
            )
            const lockfile = await sourcegraph.workspace.openTextDocument(new URL(result.uri))
            try {
                const dep = getPackageLockDependency(packageJson.text!, lockfile.text!, name)
                if (!dep) {
                    return null
                }
                return semver.satisfies(dep.version, parsedVersionRange) ? null : { packageJson, lockfile }
            } catch (err) {
                console.error(`Error checking package-lock.json and package.json for ${result.uri}.`, err, {
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
        const result = await npmExecClient({
            commands: [['npm', 'install', '--', `${dep.name}@${dep.version}`]],
            context: {
                repository: p.repoName,
                commit: p.commitID!,
                path: path.dirname(p.filePath!),
            },
            // TODO!(sqs): dir
        })
        return computeDiffs([
            { old: pkg.packageJson, newText: result.files['package.json'] },
            { old: pkg.lockfile, newText: result.files['package-lock.json'] },
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

function getPackageLockDependency(
    packageJson: string,
    packageLock: string,
    packageName: string
): { version: string } | null {
    // TODO!(sqs): this has a bug where if a package-lock.json delegates to a parent file, it throws an exception
    const tree = lockTree(JSON.parse(packageJson), JSON.parse(packageLock))
    let found: any
    // eslint-disable-next-line ban/ban
    tree.forEach((dep: any, next: any) => {
        if (dep.name === packageName) {
            found = dep
        } else {
            // eslint-disable-next-line callback-return
            next()
        }
    })
    return found ? { version: found.version } : null
}

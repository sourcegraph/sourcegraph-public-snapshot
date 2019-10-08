/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { flatten } from 'lodash'
import { from } from 'rxjs'
import { toArray } from 'rxjs/operators'
import semver from 'semver'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { PackageJsonPackageManager, ResolvedDependency, ResolvedDependencyInPackage } from '../packageManager'
import { editForDependencyUpgrade } from '../packageManagerCommon'
import { lockTree } from './logicalTree'

const npmExecClient = createExecServerClient('a8n-npm-exec')

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

        const check = async (result: sourcegraph.TextSearchResult): Promise<ResolvedDependencyInPackage | null> => {
            const packageJson = await sourcegraph.workspace.openTextDocument(
                new URL(result.uri.replace(/package-lock\.json$/, 'package.json'))
            )
            const lockfile = await sourcegraph.workspace.openTextDocument(new URL(result.uri))
            try {
                const dep = getPackageLockDependency(packageJson.text!, lockfile.text!, name)
                if (!dep) {
                    return null
                }
                return semver.satisfies(dep.version, parsedVersionRange)
                    ? null
                    : { packageJson, lockfile, dependency: dep }
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

    editForDependencyUpgrade: dep =>
        editForDependencyUpgrade(
            dep,
            [
                [
                    'npm',
                    'install',
                    '--no-audit',
                    '--package-lock-only',
                    '--ignore-scripts',
                    '--',
                    `${dep.dependency.name}@${dep.dependency.version}`,
                ],
            ],
            npmExecClient
        ),
}

function getPackageLockDependency(
    packageJson: string,
    packageLock: string,
    packageName: string
): ResolvedDependency | null {
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
    return found ? { name: packageName, version: found.version, direct: !!tree.getDep(packageName) } : null
}

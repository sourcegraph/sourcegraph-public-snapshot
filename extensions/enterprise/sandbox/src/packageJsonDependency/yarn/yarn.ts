/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { PackageJsonPackageManager, PackageJsonPackage } from '../packageManager'
import semver from 'semver'
import { flatten } from 'lodash'
import { memoizedFindTextInFiles } from '../../util'
import { from } from 'rxjs'
import { toArray } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { yarnLogicalTree } from './logicalTree'
import { editForDependencyUpgrade } from '../packageManagerCommon'

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

    editForDependencyUpgrade: (pkg, dep) =>
        editForDependencyUpgrade(
            pkg,
            dep,
            [
                [
                    'yarn',
                    'upgrade',
                    '--ignore-engines',
                    '--ignore-platform',
                    '--ignore-scripts',
                    '--non-interactive',
                    '--no-node-version-check',
                    '--no-progress',
                    '--silent',
                    '--skip-integrity-check',
                    '--no-default-rc',
                    '--',
                    `${dep.name}@${dep.version}`,
                ],
            ],
            yarnExecClient
        ),
}

function getYarnLockDependency(packageJson: string, yarnLock: string, packageName: string): { version: string } | null {
    const tree = yarnLogicalTree(JSON.parse(packageJson), yarnLock)
    let found: any
    // eslint-disable-next-line ban/ban
    tree.forEach((dep: { name: string }, next: () => void) => {
        if (dep.name === packageName) {
            found = dep
        } else {
            // eslint-disable-next-line callback-return
            next()
        }
    })
    return found ? { version: found.version } : null
}

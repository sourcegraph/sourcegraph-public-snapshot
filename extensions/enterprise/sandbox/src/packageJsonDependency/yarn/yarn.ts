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
import { editForDependencyUpgrade, editPackageJson } from '../packageManagerCommon'
import { yarnLogicalTree } from './logicalTree'

const yarnExecClient = createExecServerClient('a8n-yarn-exec', [])

const YARN_OPTS = [
    '--ignore-engines',
    '--ignore-platform',
    '--ignore-scripts',
    '--non-interactive',
    '--no-node-version-check',
    '--no-progress',
    '--silent',
    '--mutex network',
    '--skip-integrity-check',
    '--no-default-rc',
]

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

        const check = async (result: sourcegraph.TextSearchResult): Promise<ResolvedDependencyInPackage | null> => {
            try {
                const packageJson = await sourcegraph.workspace.openTextDocument(
                    new URL(result.uri.replace(/yarn\.lock$/, 'package.json'))
                )
                const lockfile = await sourcegraph.workspace.openTextDocument(new URL(result.uri))
                try {
                    const dep = getYarnLockDependency(packageJson.text!, lockfile.text!, name)
                    if (!dep) {
                        return null
                    }
                    return semver.satisfies(dep.version, parsedVersionRange)
                        ? null
                        : { packageJson, lockfile, dependency: dep }
                } catch (err) {
                    console.error(`Error checking yarn.lock and package.json for ${result.uri}.`, err, {
                        lockfile: lockfile.text,
                        packagejson: packageJson.text,
                    })
                    return null
                }
            } catch (err) {
                console.error(`Error getting yarn.lock and package.json for ${result.uri}`, err)
                return null
            }
        }
        return (await Promise.all(results.map(check))).filter(isDefined)
    },

    editForDependencyUpgrade: async dep => {
        if (dep.dependency.direct) {
            return editForDependencyUpgrade(
                dep,
                [['yarn', 'upgrade', ...YARN_OPTS, '--', `${dep.dependency.name}@${dep.dependency.version}`]],
                yarnExecClient
            )
        }

        // Handle indirect dep upgrades by adding to Yarn resolutions. This causes an error in `yarn
        // check` if the resolution is not compatible. TODO(sqs): Find the minimum upgrade path (if
        // any) to eliminate the old version dep without using resolutions.
        const workspaceEdit = editPackageJson(dep.packageJson, [
            { path: ['resolutions', dep.dependency.name], value: dep.dependency.version },
        ])
        const packageJsonObj = JSON.parse(dep.packageJson.text!)
        const edits2 = await editForDependencyUpgrade(
            {
                ...dep,
                packageJson: {
                    uri: dep.packageJson.uri,
                    text: JSON.stringify({
                        ...packageJsonObj,
                        resolutions: { ...packageJsonObj.resolutions, [dep.dependency.name]: dep.dependency.version },
                    }),
                },
            },
            [['yarn', ...YARN_OPTS, 'install']],
            yarnExecClient
        )
        workspaceEdit.set(new URL(dep.lockfile.uri), edits2.get(new URL(dep.lockfile.uri)))
        return workspaceEdit
    },
}

function getYarnLockDependency(packageJson: string, yarnLock: string, packageName: string): ResolvedDependency | null {
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
    return found ? { name: packageName, version: found.version, direct: !!tree.getDep(packageName) } : null
}

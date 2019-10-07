import { PackageJsonPackageManager, PackageJsonPackage } from './packageManager'
import semver from 'semver'
import logicalTree from 'npm-logical-tree'
import { flatten } from 'lodash'
import { memoizedFindTextInFiles } from '../util'
import { from } from 'rxjs'
import { toArray } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../shared/src/util/types'

interface PackageLockJson {
	lockfileVersion: number
	dependencies: {[name: string]: {version: string}}
}

export const npmPackageManager: PackageJsonPackageManager = {
    packagesWithUnsatisfiedDependencyVersionRange: async (name, versionRange) => {
			const parsedVersionRange = new semver.Range(versionRange)

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
					const packageJson = await sourcegraph.workspace.openTextDocument(new URL(result.uri.replace(/package-lock\.json$/, 'package.json')))
					const lockfile = await sourcegraph.workspace.openTextDocument(new URL(result.uri))
					const tree = logicalTree(packageJson.text!, lockfile.text!)
					const dep = tree.getDep(name)
					return semver.satisfies(dep.version, parsedVersionRange) ? null : {packageJson,lockfile}
				}
				return (await Promise.all(results.map(check))).filter(isDefined)
    },
}

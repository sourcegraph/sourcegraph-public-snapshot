/* eslint-disable @typescript-eslint/no-non-null-assertion */
import semver from 'semver'
import { PackageJsonDependencyQuery, ResolvedDependency } from './packageManager'
import { DependencySpecification } from '../dependencyManagement'
import * as sourcegraph from 'sourcegraph'
import { Observable, combineLatest } from 'rxjs'
import { openTextDocument } from '../dependencyManagement/util'
import { map } from 'rxjs/operators'

export const provideDependencySpecification = (
    packageJson: URL,
    lockfile: URL,
    query: PackageJsonDependencyQuery & { parsedVersionRange: semver.Range },
    getDependencyFromPackageJsonAndLockfile: (
        packageJson: string,
        lockfile: string,
        depName: string
    ) => ResolvedDependency | null
): Observable<DependencySpecification<PackageJsonDependencyQuery> | null> =>
    combineLatest([openTextDocument(packageJson), openTextDocument(lockfile)]).pipe(
        map(([packageJson, lockfile]) => {
            if (packageJson === null || lockfile === null) {
                return null
            }
            const partialSpec: Pick<DependencySpecification<PackageJsonDependencyQuery>, 'query'> = {
                query,
            }
            try {
                // TODO!(sqs): support multiple versions in lockfile/package.json
                const dep = getDependencyFromPackageJsonAndLockfile(packageJson.text!, lockfile.text!, query.name)
                if (!dep) {
                    return null
                }
                if (!semver.satisfies(dep.version, query.parsedVersionRange)) {
                    return null
                }
                const spec: DependencySpecification<PackageJsonDependencyQuery> = {
                    ...partialSpec,
                    declarations: [
                        {
                            name: dep.name,
                            // requestedVersion: // TODO!(sqs): get from package.json
                            direct: dep.direct,
                            location: {
                                uri: new URL(packageJson.uri),
                                // TODO!(sqs): this is not exact anyway, needs to traverse json file
                                range: findMatchRange(packageJson.text!, `"${query.name}"`),
                            },
                        },
                    ],
                    resolutions: [
                        {
                            name: dep.name,
                            resolvedVersion: dep.version,
                            location: {
                                uri: new URL(lockfile.uri),
                                // TODO!(sqs): this differs from yarn.lock vs package-lock.json and is not exact anyway, needs to traverse json file
                                range: findMatchRange(packageJson.text!, query.name),
                            },
                        },
                    ],
                }
                return spec
            } catch (err) {
                const spec: DependencySpecification<PackageJsonDependencyQuery> = {
                    ...partialSpec,
                    declarations: [],
                    resolutions: [],
                    error: Object.assign(
                        new Error(
                            `Unable to get dependency specification from package.json and lockfile (package ${JSON.stringify(
                                query.name
                            )} in ${packageJson.uri}): ${err.message}`
                        ),
                        { data: { packageJson: packageJson.uri, query }, wrapped: err }
                    ),
                }
                return spec
            }
        })
    )

function findMatchRange(text: string, str: string): sourcegraph.Range | undefined {
    for (const [i, line] of text.split('\n').entries()) {
        const j = line.indexOf(str)
        if (j !== -1) {
            return new sourcegraph.Range(i, j, i, j + str.length)
        }
    }
    return undefined
}

/* eslint-disable @typescript-eslint/no-non-null-assertion */
import semver from 'semver'
import { PackageJsonDependencyQuery, ResolvedDependency } from './packageManager'
import { DependencySpecification } from '../dependencyManagement'
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
        map(([packageJson, yarnLock]) => {
            if (packageJson === null || yarnLock === null) {
                return null
            }
            try {
                // TODO!(sqs): support multiple versions in lockfile/package.json
                const dep = getDependencyFromPackageJsonAndLockfile(packageJson.text!, yarnLock.text!, name)
                if (!dep) {
                    return null
                }
                if (!semver.satisfies(dep.version, query.parsedVersionRange)) {
                    return null
                }
                const spec: DependencySpecification<PackageJsonDependencyQuery> = {
                    query,
                    declarations: [
                        {
                            name: dep.name,
                            // requestedVersion: // TODO!(sqs): get from package.json
                            direct: dep.direct,
                            location: { uri: new URL(packageJson.uri) },
                        },
                    ],
                    resolutions: [
                        {
                            name: dep.name,
                            resolvedVersion: dep.version,
                            location: { uri: new URL(yarnLock.uri) },
                        },
                    ],
                }
                return spec
            } catch (err) {
                const err2 = new Error('Unable to get dependency specification from package.json and lockfile')
                ;(err2 as any).data = { packageJson: packageJson.toString(), query }
                err.wrapped = err
                throw err2
            }
        })
    )

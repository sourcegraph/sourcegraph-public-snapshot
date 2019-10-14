/* eslint-disable @typescript-eslint/no-non-null-assertion */
import { from, merge, Observable, combineLatest, of } from 'rxjs'
import { toArray, switchMap, filter, map } from 'rxjs/operators'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { JavaDependencyManagementProvider, JavaDependencyQuery } from '..'
import { DependencySpecification, DependencyDeclaration, DependencyResolution } from '../../dependencyManagement'
import { editForCommands } from '../../execServer/editsForCommands'
import { openTextDocument, findMatchRange } from '../../dependencyManagement/util'
import { parseDependenciesLock } from './dependenciesLock'
import { Location } from 'sourcegraph'
import { parseDependencyNotation } from './util'
import { replaceVersion } from './replaceDependencyVersion'

const gradleExecClient = createExecServerClient('a8n-java-gradle-exec')

const GRADLE_OPTS = ['--no-daemon']

const provideDependencySpecification = (
    buildGradle: URL,
    dependenciesLock: URL,
    query: JavaDependencyQuery
): Observable<DependencySpecification<JavaDependencyQuery> | null> =>
    combineLatest([openTextDocument(buildGradle), openTextDocument(dependenciesLock)]).pipe(
        switchMap(([buildGradle, dependenciesLock]) => {
            if (!buildGradle) {
                throw new Error('Unable to fetch build.gradle.')
            }
            if (!dependenciesLock) {
                console.warn(`Ignoring build.gradle file with no corresponding dependency.lock: ${buildGradle.uri}`)
                return of(null)
            }

            const declarations: DependencyDeclaration[] = []
            const resolutions: DependencyResolution[] = []

            // TODO!(sqs): handle multiple/differing dep versions per configuration
            const seenDeps = new Set<string>()
            const lockedDeps = parseDependenciesLock(dependenciesLock.text!)
            for (const [, deps] of Object.entries(lockedDeps)) {
                for (const [id, { locked, requested }] of Object.entries(deps)) {
                    if (seenDeps.has(id)) {
                        continue
                    }
                    seenDeps.add(id)

                    const declRange = findMatchRange(buildGradle.text!, id)
                    const lockfileLocation: Location = {
                        uri: new URL(dependenciesLock.uri),
                        range: findMatchRange(dependenciesLock.text!, id),
                    }
                    declarations.push({
                        name: id,
                        requestedVersion: requested,
                        location: declRange ? { uri: new URL(buildGradle.uri), range: declRange } : lockfileLocation,
                        direct: true, // TODO!(sqs),
                    })
                    resolutions.push({
                        name: id,
                        location: lockfileLocation,
                        resolvedVersion: locked,
                    })
                }
            }

            const spec: DependencySpecification<JavaDependencyQuery> = {
                query,
                declarations,
                resolutions,
            }
            return of(spec)
        })
    )

export const gradleDependencyManagementProvider: JavaDependencyManagementProvider = {
    type: 'gradle',
    provideDependencySpecifications: (query, filters = '') =>
        from(
            memoizedFindTextInFiles(
                {
                    pattern: `${JSON.stringify(query.name)} ${filters}`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)build.gradle$'],
                        excludes: ['node_modules'],
                        type: 'regexp',
                    },
                    maxResults: 99999999, // TODO!(sqs): un-hardcode
                }
            )
        ).pipe(
            switchMap(textSearchResults =>
                merge(
                    ...textSearchResults.map(textSearchResult =>
                        provideDependencySpecification(
                            new URL(textSearchResult.uri),
                            new URL(textSearchResult.uri.replace(/build\.gradle$/, 'dependencies.lock')),
                            query
                        )
                    )
                ).pipe(
                    filter(isDefined),
                    toArray()
                )
            )
        ),
    resolveDependencyUpgradeAction: (dep, version) => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length === 0) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        const decl = dep.declarations[0]
        const res = dep.resolutions[0]
        if (!res.location) {
            throw new Error('invalid lockfile with no match location')
        }
        const newBuildGradle = openTextDocument(decl.location.uri).pipe(
            map(buildGradle => {
                const dependencyNotation = parseDependencyNotation(decl.name)
                return replaceVersion(buildGradle!.text!, {
                    group: dependencyNotation.group,
                    name: dependencyNotation.name,
                    oldVersion: decl.requestedVersion!,
                    newVersion: version,
                })
            })
        )
        return newBuildGradle.pipe(
            switchMap(newBuildGradle =>
                editForCommands(
                    [
                        { uri: decl.location.uri.toString(), text: newBuildGradle },
                        ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
                    ],
                    [
                        [
                            './gradlew',
                            ...GRADLE_OPTS,
                            'generateLock',
                            'saveLock',
                            '-PdependencyLock.includeTransitives=true',
                        ],
                    ],
                    gradleExecClient
                )
            )
        )
    },
    resolveDependencyBanAction: () => {
        throw new Error('banning gradle dependencies is not yet supported')
    },
}

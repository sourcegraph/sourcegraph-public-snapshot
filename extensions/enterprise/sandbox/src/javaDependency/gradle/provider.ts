/* eslint-disable @typescript-eslint/no-non-null-assertion */
import semver from 'semver'
import { from, merge, Observable, combineLatest, of, throwError } from 'rxjs'
import { toArray, switchMap, filter, map, tap, share, catchError, first } from 'rxjs/operators'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { JavaDependencyManagementProvider, JavaDependencyQuery } from '..'
import { DependencySpecification, DependencyDeclaration, DependencyResolution } from '../../dependencyManagement'
import { editForCommands } from '../../execServer/editsForCommands'
import { openTextDocument, findMatchRange } from '../../dependencyManagement/util'
import { parseDependenciesLock } from './dependenciesLock'
import { Location, Range, Position, WorkspaceEdit, DiagnosticSeverity } from 'sourcegraph'
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
                const declRange = findMatchRange(buildGradle.text!, query.name)
                return of<DependencySpecification<JavaDependencyQuery>>({
                    query,
                    declarations: declRange
                        ? [
                              {
                                  name: query.name,
                                  requestedVersion: query.versionRange,
                                  direct: true,
                                  location: { uri: new URL(buildGradle.uri), range: declRange },
                              },
                          ]
                        : [],
                    resolutions: [],
                    messages: [`Ignoring build.gradle file with no corresponding dependency.lock: ${buildGradle.uri}`],
                })
            }

            const declarations: DependencyDeclaration[] = []
            const resolutions: DependencyResolution[] = []

            // TODO!(sqs): handle multiple/differing dep versions per configuration
            const seenDeps = new Set<string>()
            const lockedDeps = parseDependenciesLock(dependenciesLock.text!, dependenciesLock.uri)
            for (const [, deps] of Object.entries(lockedDeps)) {
                for (const [id, { locked, requested }] of Object.entries(deps)) {
                    // TODO!(sqs): handle semver satisfies for both requested and locked?
                    if (id !== query.name) {
                        continue
                    }
                    if (
                        !semver.satisfies(semver.coerce(requested) || requested, query.parsedVersionRange) &&
                        requested !== query.versionRange
                    ) {
                        continue
                    }

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
                    pattern: `${JSON.stringify(query.name)} ${filters} index:only`,
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
                ).pipe(filter(isDefined))
            ),
            toArray()
        ),
    resolveDependencyUpgradeAction: (dep, version) => {
        // TODO!(sqs): this is not correct w.r.t. indirect deps
        if (dep.declarations.length === 0) {
            console.error('Invalid declarations.', dep)
            throw new Error('invalid declarations')
        }
        const decl = dep.declarations[0]
        // const res = dep.resolutions[0]
        // if (!res || !res.location) {
        //     throw new Error('invalid lockfile with no match location')
        // }
        const existingBuildGradle = openTextDocument(decl.location.uri).pipe(
            switchMap(doc => {
                if (!doc) {
                    return throwError('no build.gradle')
                }
                return of(doc)
            }),
            share()
        )
        const newBuildGradle = existingBuildGradle.pipe(
            map(buildGradle => {
                const dependencyNotation = parseDependencyNotation(decl.name)
                return replaceVersion(buildGradle.text!, {
                    group: dependencyNotation.group,
                    name: dependencyNotation.name,
                    oldVersion: decl.requestedVersion!,
                    newVersion: version,
                })
            })
        )
        const lockfileDiff =
            dep.resolutions.length > 0
                ? newBuildGradle.pipe(
                      switchMap(newBuildGradle =>
                          editForCommands(
                              [
                                  { uri: decl.location.uri.toString(), text: newBuildGradle },
                                  ...dep.resolutions.filter(propertyIsDefined('location')).map(d => d.location.uri),
                              ],
                              [
                                  [
                                      'gradle',
                                      ...GRADLE_OPTS,
                                      'generateLock',
                                      'saveLock',
                                      '-PdependencyLock.includeTransitives=true',
                                  ],
                              ],
                              gradleExecClient
                          ).pipe(
                              catchError(err => {
                                  const edit = new WorkspaceEdit()
                                  edit.addDiagnostic({
                                      resource: decl.location.uri,
                                      range: new Range(0, 0, 0, 0),
                                      severity: DiagnosticSeverity.Warning,
                                      message: 'Error regenerating dependencies.lock (`gradle generateLock saveLock`)',
                                      detail:
                                          '```text\n' +
                                          (err.message as string).replace(/^editForCommands.*$/m, '').trim() +
                                          '\n```',
                                  })
                                  return [edit]
                              })
                          )
                      )
                  )
                : of<WorkspaceEdit>(new WorkspaceEdit())
        return combineLatest([existingBuildGradle, newBuildGradle, lockfileDiff]).pipe(
            map(([existingBuildGradle, newBuildGradle, edit]) => {
                // Also add build.gradle edit.
                edit.replace(
                    new URL(existingBuildGradle.uri),
                    new Range(new Position(0, 0), existingBuildGradle.positionAt(existingBuildGradle.text!.length)),
                    newBuildGradle
                )
                return edit
            })
        )
    },
    resolveDependencyBanAction: () => {
        throw new Error('banning gradle dependencies is not yet supported')
    },
}

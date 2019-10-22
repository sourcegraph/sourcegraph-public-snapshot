/* eslint-disable @typescript-eslint/no-non-null-assertion */
import semver from 'semver'
import { flatten } from 'lodash'
import { from, merge, Observable, combineLatest, of, throwError } from 'rxjs'
import { toArray, switchMap, filter, map, share, catchError, startWith, tap, first } from 'rxjs/operators'
import { isDefined, propertyIsDefined } from '../../../../../../shared/src/util/types'
import { createExecServerClient } from '../../execServer/client'
import { memoizedFindTextInFiles } from '../../util'
import { JavaDependencyManagementProvider, JavaDependencyQuery, JavaDependencyCampaignContext } from '..'
import { DependencySpecification, DependencyDeclaration, DependencyResolution } from '../../dependencyManagement'
import { editForCommands } from '../../execServer/editsForCommands'
import { openTextDocument, findMatchRange } from '../../dependencyManagement/util'
import { parseDependenciesLock } from './dependenciesLock'
import { Location, Range, Position, WorkspaceEdit, DiagnosticSeverity } from 'sourcegraph'
import { parseDependencyNotation } from './util'
import { replaceVersion } from './replaceDependencyVersion'
import { parseRepoURI } from '../../../../../../shared/src/util/url'
import { dirname } from 'path'
import { LOADING } from '../../dependencyManagement/common'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'

const gradleExecClient = createExecServerClient('a8n-java-gradle-exec')

const GRADLE_OPTS = ['--no-daemon']

const hasDependency = async (buildGradle: URL, query: JavaDependencyQuery): Promise<boolean> => {
    const p = parseRepoURI(buildGradle.toString())
    const result = await gradleExecClient({
        commands: [
            ['gradle', ...GRADLE_OPTS, 'dependencyInsight', '--dependency', `${query.name}:${query.versionRange}`],
        ],
        context: {
            repository: p.repoName,
            commit: p.commitID!,
        },
        dir: dirname(p.filePath!),
        label: `dependencyInsight(${buildGradle.toString()})`,
    })
    const commandResult = result.commands[0]
    if (!commandResult.ok) {
        throw new Error(`gradle dependencyInsight on ${buildGradle.toString()} failed: ${commandResult.error}`)
    }
    return !commandResult.combinedOutput.includes('No dependencies matching given input were found')
}

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
                // if (!query.supportMissingDependencyLock) {
                //     return of(null)
                // }

                // Use `gradle dependencyInsight` to determine if this dependency exists
                // transitively.
                return from(
                    query.supportMissingDependencyLock ? hasDependency(new URL(buildGradle.uri), query) : of(false)
                ).pipe(
                    query.supportMissingDependencyLock ? startWith(LOADING) : tap(),
                    catchError(err => of<ErrorLike>(asError(err))),
                    map(hasDependency => ({
                        query,
                        declarations:
                            declRange && hasDependency !== LOADING && !isErrorLike(hasDependency)
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
                        diagnostics: [
                            {
                                resource: new URL(buildGradle.uri),
                                range: declRange || new Range(0, 0, 0, 0),
                                ...(isErrorLike(hasDependency)
                                    ? {
                                          message: 'No dependency.lock found and `gradle dependencyInsight` failed',
                                          detail:
                                              '```text\n' +
                                              (hasDependency.message as string)
                                                  .replace(/^dependencyInsight\($/m, '')
                                                  .trim() +
                                              '\n```',
                                          severity: DiagnosticSeverity.Error,
                                      }
                                    : query.supportMissingDependencyLock
                                    ? {
                                          message: `No dependency.lock found (${
                                              hasDependency === LOADING ? 'using' : 'used'
                                          } slower gradle dependencyInsight)`,
                                          severity: DiagnosticSeverity.Hint,
                                      }
                                    : {
                                          message: `No dependency.lock found`,
                                          detail:
                                              'Set variable `supportMissingDependencyLock` to `true` to run `gradle dependencyInsight` to compute the transitive dependency set on-the-fly.',
                                          severity: DiagnosticSeverity.Warning,
                                      }),
                            },
                        ],
                        isLoading: hasDependency === LOADING,
                    }))
                )
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
                    pattern: `${filters} index:only`,
                    type: 'regexp',
                },
                {
                    repositories: {
                        type: 'regexp',
                    },
                    files: {
                        includes: ['(^|/)build.gradle$'],
                        type: 'regexp',
                    },
                    maxResults: 100, // TODO!(sqs): un-hardcode
                }
            )
        ).pipe(
            toArray(),
            map(textSearchResults => flatten(textSearchResults)),
            switchMap(textSearchResults =>
                combineLatest(
                    ...textSearchResults.map(textSearchResult =>
                        provideDependencySpecification(
                            new URL(textSearchResult.uri),
                            new URL(textSearchResult.uri.replace(/build\.gradle$/, 'dependencies.lock')),
                            query
                        )
                    )
                ).pipe(filter(isDefined))
            )
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
                if (existingBuildGradle.text !== newBuildGradle) {
                    // Also add build.gradle edit.
                    edit.replace(
                        new URL(existingBuildGradle.uri),
                        new Range(new Position(0, 0), existingBuildGradle.positionAt(existingBuildGradle.text!.length)),
                        newBuildGradle
                    )
                }
                return edit
            })
        )
    },
    resolveDependencyBanAction: () => {
        throw new Error('banning gradle dependencies is not yet supported')
    },
}

// TODO!(sqs): https://github.com/kevcodez/gradle-upgrade-interactive/blob/master/ReplaceVersion.js

import { from, Observable, of, Subscription, Unsubscribable } from 'rxjs'
import semver from 'semver'
import { filter, map, startWith, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../shared/src/util/types'
import { javaDependencyManagementProviderRegistry, JavaDependencyQuery } from '.'
import { DependencySpecificationWithType } from '../dependencyManagement/combinedProvider'
import { toLocation } from '../../../../../shared/src/api/extension/api/types'

const COMMAND_ID = 'javaDependency.action'

export interface JavaDependencyCampaignContext {
    packageName?: string
    matchVersion?: string
    action: JavaDependencyAction
    createChangesets: boolean
    filters?: string
}

export type JavaDependencyAction = { requireVersion: string }

const LOADING = 'loading' as const

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('javaDependency', {
            provideDiagnostics: (_scope, context) =>
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                provideDiagnostics((context as any) as JavaDependencyCampaignContext).pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(COMMAND_ID, diagnostic => {
            if (!diagnostic || (diagnostic.tags && !diagnostic.tags.includes('fix'))) {
                return Promise.resolve(new sourcegraph.WorkspaceEdit())
            }
            return editForDependencyAction(diagnostic).toPromise()
        })
    )
    return subscriptions
}

const DEPENDENCY_TAG = 'type:javaDependency'

interface DiagnosticData extends DependencySpecificationWithType<JavaDependencyQuery> {
    action: JavaDependencyCampaignContext['action']
}

function provideDiagnostics({
    packageName,
    matchVersion,
    action,
    createChangesets,
    filters,
}: JavaDependencyCampaignContext): Observable<sourcegraph.Diagnostic[] | typeof LOADING> {
    return packageName && matchVersion && action
        ? from(sourcegraph.workspace.rootChanges).pipe(
              startWith(undefined),
              map(() => sourcegraph.workspace.roots),
              switchMap(roots => {
                  if (roots.length > 0) {
                      return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
                  }

                  const depQuery: JavaDependencyQuery = {
                      name: packageName,
                      versionRange: matchVersion,
                      parsedVersionRange: new semver.Range(matchVersion),
                  }
                  const specs = javaDependencyManagementProviderRegistry.provideDependencySpecifications(
                      depQuery,
                      filters
                  )
                  return specs.pipe(
                      map(specs =>
                          specs
                              .map(spec => {
                                  if (spec.error) {
                                      console.error(spec.error)
                                      return null
                                  }
                                  const specMain = spec.declarations[0]
                                      ? spec.declarations[0]
                                      : { ...spec.resolutions[0], direct: false }
                                  if (!specMain.location) {
                                      return null
                                  }
                                  const data: DiagnosticData = { ...spec, action }
                                  const diagnostic: sourcegraph.Diagnostic = {
                                      resource: specMain.location.uri,
                                      message: `${specMain.direct ? '' : 'Indirect '}npm dependency ${specMain.name}${
                                          depQuery.versionRange === '*' ? '' : `@${depQuery.versionRange}`
                                      } ${
                                          action === 'ban'
                                              ? 'is banned'
                                              : `must be upgraded to ${action.requireVersion}`
                                      }`,
                                      range: specMain.location.range || new sourcegraph.Range(0, 0, 0, 0),
                                      severity: sourcegraph.DiagnosticSeverity.Warning,
                                      // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                                      data: JSON.stringify(data),
                                      tags: [DEPENDENCY_TAG, packageName, createChangesets ? 'fix' : undefined].filter(
                                          isDefined
                                      ),
                                  }
                                  return diagnostic
                              })
                              .filter(isDefined)
                      )
                  )
              }),
              startWith(LOADING)
          )
        : of<sourcegraph.Diagnostic[]>([])
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (_doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(d => isProviderDiagnostic(d) && d.tags && d.tags.includes('fix'))
            if (!diag) {
                return of<sourcegraph.Action[]>([])
            }
            return editForDependencyAction(diag).pipe(
                map(edit => [
                    {
                        title: 'Upgrade dependency in package.json',
                        edit,
                        computeEdit: { title: 'Upgrade dependency', command: COMMAND_ID },
                        diagnostics: [diag],
                    },
                ])
            )
        },
    }
}

function isProviderDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return !!diag.tags && diag.tags.includes(DEPENDENCY_TAG)
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): DiagnosticData {
    if (!diag.data) {
        throw new Error('no diagnostic data')
    }
    const parsed: DiagnosticData = JSON.parse(diag.data)
    return {
        ...parsed,
        declarations: parsed.declarations.map(d => ({ ...d, location: toLocation(d.location as any) })),
        resolutions: parsed.resolutions.map(r => ({
            ...r,
            location: r.location ? toLocation(r.location as any) : undefined,
        })),
    }
}

function editForDependencyAction(diag: sourcegraph.Diagnostic): Observable<sourcegraph.WorkspaceEdit> {
    const data = getDiagnosticData(diag)
    if (data.action === 'ban') {
        return javaDependencyManagementProviderRegistry.resolveDependencyBanAction(data)
    }
    return javaDependencyManagementProviderRegistry.resolveDependencyUpgradeAction(data, data.action.requireVersion)
}

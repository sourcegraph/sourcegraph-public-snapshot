import { from, Observable, of, Subscription, Unsubscribable } from 'rxjs'
import { filter, map, startWith, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../shared/src/util/types'
import { PackageJsonDependencyQuery } from './packageManager'
import { packageJsonDependencyManagementProviderRegistry } from './providers'
import { DependencySpecificationWithType } from '../dependencyManagement/combinedProvider'

const COMMAND_ID = 'packageJsonDependency.action'

export interface PackageJsonDependencyCampaignContext {
    packageName?: string
    matchVersion?: string
    action: PackageJsonDependencyAction
    createChangesets: boolean
    filters?: string
}

export type PackageJsonDependencyAction = { requireVersion: string } | 'ban'

const LOADING = 'loading' as const

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('packageJsonDependency', {
            provideDiagnostics: (_scope, context) =>
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                provideDiagnostics((context as any) as PackageJsonDependencyCampaignContext).pipe(
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

const DEPENDENCY_TAG = 'type:packageJsonDependency'

interface DiagnosticData extends DependencySpecificationWithType<PackageJsonDependencyQuery> {
    action: PackageJsonDependencyCampaignContext['action']
}

function provideDiagnostics({
    packageName,
    matchVersion,
    action,
    createChangesets,
    filters,
}: PackageJsonDependencyCampaignContext): Observable<sourcegraph.Diagnostic[] | typeof LOADING> {
    return packageName && matchVersion && action
        ? from(sourcegraph.workspace.rootChanges).pipe(
              startWith(undefined),
              map(() => sourcegraph.workspace.roots),
              switchMap(roots => {
                  if (roots.length > 0) {
                      return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
                  }

                  const depQuery: PackageJsonDependencyQuery = {
                      name: packageName,
                      versionRange: matchVersion,
                  }
                  const specs = packageJsonDependencyManagementProviderRegistry.provideDependencySpecifications(
                      depQuery,
                      filters
                  )
                  return specs.pipe(
                      map(specs =>
                          specs.map(spec => {
                              const mainDecl = spec.declarations[0]
                              if (!mainDecl.location.range) {
                                  throw new Error('no range')
                              }
                              const data: DiagnosticData = { ...spec, action }
                              const diagnostic: sourcegraph.Diagnostic = {
                                  resource: mainDecl.location.uri,
                                  message: `${mainDecl.direct ? '' : 'Indirect '}npm dependency ${mainDecl.name}${
                                      depQuery.versionRange === '*' ? '' : `@${depQuery.versionRange}`
                                  } ${action === 'ban' ? 'is banned' : `must be upgraded to ${action.requireVersion}`}`,
                                  range: mainDecl.location.range,
                                  severity: sourcegraph.DiagnosticSeverity.Warning,
                                  // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                                  data: JSON.stringify(data),
                                  tags: [DEPENDENCY_TAG, packageName, createChangesets ? 'fix' : undefined].filter(
                                      isDefined
                                  ),
                              }
                              return diagnostic
                          })
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
    return JSON.parse(diag.data)
}

function editForDependencyAction(diag: sourcegraph.Diagnostic): Observable<sourcegraph.WorkspaceEdit> {
    const data = getDiagnosticData(diag)
    if (data.action === 'ban') {
        return packageJsonDependencyManagementProviderRegistry.resolveDependencyBanAction(data)
    }
    return packageJsonDependencyManagementProviderRegistry.resolveDependencyUpgradeAction(
        data,
        data.action.requireVersion
    )
}

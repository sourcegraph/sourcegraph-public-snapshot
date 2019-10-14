import { flatten } from 'lodash'
import { from, Observable, of, Subscription, Unsubscribable } from 'rxjs'
import { filter, map, startWith, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../shared/src/util/types'
import { npmPackageManager } from './npm/npm'
import { PackageJsonDependency, ResolvedDependency, PackageJsonDependencyQuery } from './packageManager'
import { yarnPackageManager } from './yarn/yarn'
import { packageJsonDependencyManagementProviderRegistry } from './providers'

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
                provideDiagnostics((context as any) as PackageJsonDependencyCampaignContext).pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(COMMAND_ID, diagnostic => {
            if (!diagnostic || (diagnostic.tags && !diagnostic.tags.includes('fix'))) {
                return new sourcegraph.WorkspaceEdit()
            }
            return editForDependencyAction(diagnostic)
        })
    )
    return subscriptions
}

const DEPENDENCY_TAG = 'type:packageJsonDependency'

interface DiagnosticData {
    dependency: ResolvedDependency
    packageJson: { uri: string; text: string }
    lockfile: { uri: string; text: string }
    action: PackageJsonDependencyCampaignContext['action']
    type: 'npm' | 'yarn'
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
              switchMap(async roots => {
                  if (roots.length > 0) {
                      return [] as sourcegraph.Diagnostic[] // TODO!(sqs): dont run in comparison mode
                  }

                  const depQuery: PackageJsonDependencyQuery = {
                      name: packageName,
                      versionRange: matchVersion,
                  }
                  const specs = packageJsonDependencyManagementProviderRegistry.provideDependencySpecifications(
                      depQuery,
                      filters
                  )
                  return flatten(
                      hits
                          .map(({ type, ...hit }) => {
                              let matchRange = findMatchRange(hit.packageJson.text!, `"${packageName}"`)
                              let matchDoc: sourcegraph.TextDocument | undefined
                              if (matchRange) {
                                  matchDoc = hit.packageJson
                              }
                              if (!matchRange) {
                                  matchRange = findMatchRange(
                                      hit.lockfile.text!,
                                      type === 'npm' ? `"${packageName}"` : `${packageName}@`
                                  )
                                  if (matchRange) {
                                      matchDoc = hit.lockfile
                                  }
                              }

                              if (!matchRange || !matchDoc) {
                                  return null
                              }

                              const diagnostic: sourcegraph.Diagnostic = {
                                  resource: new URL(matchDoc.uri),
                                  message: `${
                                      matchDoc === hit.lockfile ? 'Indirect ' : ''
                                  }npm dependency ${packageName}${matchVersion === '*' ? '' : `@${matchVersion}`} ${
                                      action === 'ban' ? 'is banned' : `must be upgraded to ${action.requireVersion}`
                                  }`,
                                  range: matchRange,
                                  severity: sourcegraph.DiagnosticSeverity.Warning,
                                  // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                                  data: JSON.stringify({
                                      dependency: hit.dependency,
                                      packageJson: { uri: hit.packageJson.uri },
                                      lockfile: { uri: hit.lockfile.uri },
                                      action,
                                      type,
                                  } as DiagnosticData),
                                  tags: [DEPENDENCY_TAG, packageName, createChangesets ? 'fix' : undefined].filter(
                                      isDefined
                                  ),
                              }
                              return [diagnostic]
                          })
                          .filter(isDefined)
                  )
              }),
              startWith(LOADING)
          )
        : of<sourcegraph.Diagnostic[]>([])
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: async (_doc, _rangeOrSelection, context): Promise<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(d => isProviderDiagnostic(d) && d.tags && d.tags.includes('fix'))
            if (!diag) {
                return []
            }
            return [
                {
                    title: 'Upgrade dependency in package.json',
                    edit: await editForDependencyAction(diag),
                    computeEdit: { title: 'Upgrade dependency', command: COMMAND_ID },
                    diagnostics: [diag],
                },
            ]
        },
    }
}

function findMatchRange(text: string, str: string): sourcegraph.Range | null {
    for (const [i, line] of text.split('\n').entries()) {
        const j = line.indexOf(str)
        if (j !== -1) {
            return new sourcegraph.Range(i, j, i, j + str.length)
        }
    }
    return null
}

function isProviderDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return !!diag.tags && diag.tags.includes(DEPENDENCY_TAG)
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): DiagnosticData {
    return JSON.parse(diag.data!)
}

async function editForDependencyAction(diag: sourcegraph.Diagnostic): Promise<sourcegraph.WorkspaceEdit> {
    const data = getDiagnosticData(diag)
    return await (data.type === 'npm' ? npmPackageManager : yarnPackageManager).editForDependencyAction(
        {
            packageJson: await sourcegraph.workspace.openTextDocument(new URL(data.packageJson.uri)),
            lockfile: await sourcegraph.workspace.openTextDocument(new URL(data.lockfile.uri)),
            dependency: data.dependency,
        },
        data.action
    )
}

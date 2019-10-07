import * as sourcegraph from 'sourcegraph'
import { flatten } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from } from 'rxjs'
import { map, switchMap, startWith, filter } from 'rxjs/operators'
import { isDefined } from '../../../../../shared/src/util/types'
import { npmPackageManager } from './npm/npm'

const UPGRADE_DEPENDENCY_COMMAND = 'packageJsonDependency.upgrade'

export interface PackageJsonDependencyCampaignContext {
    packageName?: string
    upgradeToVersion?: string
    createChangesets: boolean
    filters?: string
}

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
        sourcegraph.commands.registerActionEditCommand(UPGRADE_DEPENDENCY_COMMAND, async diagnostic => {
            if (!diagnostic) {
                return new sourcegraph.WorkspaceEdit()
            }
            const doc = await sourcegraph.workspace.openTextDocument(diagnostic.resource)
            return computeUpgradeDependencyEdit(diagnostic, doc)
        })
    )
    return subscriptions
}

const DEPENDENCY_TAG = 'type:packageJsonDependency'

interface DiagnosticData {
    dependency: Pick<PackageJsonDependencyCampaignContext, 'packageName' | 'upgradeToVersion'>
    packageJson: { uri: string; text: string }
    lockfile: { uri: string; text: string }
}

function provideDiagnostics({
    packageName,
    upgradeToVersion,
}: PackageJsonDependencyCampaignContext): Observable<sourcegraph.Diagnostic[] | typeof LOADING> {
    return packageName && upgradeToVersion
        ? from(sourcegraph.workspace.rootChanges).pipe(
              startWith(undefined),
              map(() => sourcegraph.workspace.roots),
              switchMap(async roots => {
                  if (roots.length > 0) {
                      return [] as sourcegraph.Diagnostic[] // TODO!(sqs): dont run in comparison mode
                  }

                  const hits = await npmPackageManager.packagesWithUnsatisfiedDependencyVersionRange({
                      name: packageName,
                      version: upgradeToVersion,
                  })
                  return flatten(
                      hits
                          .map(hit => {
                              const packageNameMatchString = `"${packageName}"`
                              let matchRange = findMatchRange(hit.packageJson.text!, packageNameMatchString)
                              let matchDoc: sourcegraph.TextDocument | undefined
                              if (matchRange) {
                                  matchDoc = hit.packageJson
                              }
                              if (!matchRange) {
                                  matchRange = findMatchRange(hit.lockfile.text!, packageNameMatchString)
                                  if (matchRange) {
                                      matchDoc = hit.lockfile
                                  }
                              }

                              if (!matchRange || !matchDoc) {
                                  return null
                              }

                              const diagnostic: sourcegraph.Diagnostic = {
                                  resource: new URL(matchDoc.uri),
                                  message: `npm dependency '${packageName}' must be upgraded to ${upgradeToVersion}`,
                                  range: matchRange,
                                  severity: sourcegraph.DiagnosticSeverity.Warning,
                                  // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                                  data: JSON.stringify({
                                      dependency: { packageName, upgradeToVersion },
                                      packageJson: { uri: hit.packageJson.uri, text: hit.packageJson.text },
                                      lockfile: { uri: hit.lockfile.uri, text: hit.lockfile.text },
                                  } as DiagnosticData),
                                  tags: [DEPENDENCY_TAG, packageName],
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
        provideCodeActions: async (doc, _rangeOrSelection, context): Promise<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(isProviderDiagnostic)
            if (!diag) {
                return []
            }
            const data = getDiagnosticData(diag)
            return [
                {
                    title: 'Upgrade dependency in package.json',
                    edit: await npmPackageManager.editForDependencyUpgrade(
                        {
                            packageJson: await sourcegraph.workspace.openTextDocument(new URL(data.packageJson.uri)),
                            lockfile: await sourcegraph.workspace.openTextDocument(new URL(data.lockfile.uri)),
                        },
                        { name: data.dependency.packageName!, version: data.dependency.upgradeToVersion! }
                    ),
                    computeEdit: { title: 'Upgrade dependency', command: UPGRADE_DEPENDENCY_COMMAND },
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

function computeUpgradeDependencyEdit(
    diag: sourcegraph.Diagnostic,
    doc: sourcegraph.TextDocument,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    // const data = getDiagnosticData(diag)
    // TODO!(sqs): assumes dependency key-value is all on one line and only appears once
    edit.delete(
        new URL(doc.uri),
        new sourcegraph.Range(
            diag.range.start.with({ character: 0 }),
            diag.range.end.with({ line: diag.range.end.line + 1, character: 0 })
        )
    )
    return edit
}

import * as sourcegraph from 'sourcegraph'
import { flatten, sortedUniq } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from } from 'rxjs'
import { map, switchMap, startWith, toArray, filter } from 'rxjs/operators'
import { settingsObservable, memoizedFindTextInFiles } from './util'
import { propertyIsDefined } from '../../../../shared/src/util/types'

const REMOVE_COMMAND = 'packageJsonDependency.remove'

interface Settings {}

export interface PackageJsonDependencyCampaignContext {
    packageName?: string
    versionRange?: string
    createChangesets: boolean
    showWarnings: boolean
    ban: boolean
    filters?: string
    campaignName?: string
}

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('packageJsonDependency', {
            provideDiagnostics: (_scope, context) =>
                provideDiagnostics(context).pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(REMOVE_COMMAND, async diagnostic => {
            const doc = await sourcegraph.workspace.openTextDocument(diagnostic.resource)
            return computeRemoveDependencyEdit(diagnostic, doc)
        })
    )
    return subscriptions
}

const DEPENDENCY_TAG = 'type:packageJsonDependency'

const LOADING = 'loading' as const

const provideDiagnostics = (
    context: PackageJsonDependencyCampaignContext
): Observable<sourcegraph.Diagnostic[] | typeof LOADING> =>
    context.packageName
        ? from(sourcegraph.workspace.rootChanges).pipe(
              startWith(void 0),
              map(() => sourcegraph.workspace.roots),
              switchMap(async roots => {
                  if (roots.length > 0) {
                      return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
                  }

                  const results = flatten(
                      await from(
                          memoizedFindTextInFiles(
                              {
                                  pattern: context.packageName ? globToRegExp(context.packageName).source : '',
                                  type: 'regexp',
                              },
                              {
                                  repositories: {
                                      includes: [],
                                      type: 'regexp',
                                  },
                                  files: {
                                      includes: ['(^|/)package.json$'],
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

                  const docs = await Promise.all(
                      results.map(async ({ uri }) => sourcegraph.workspace.openTextDocument(new URL(uri)))
                  )
                  return from(settingsObservable<Settings>()).pipe(
                      map(() =>
                          flatten(
                              docs
                                  .filter(doc => doc.text.length < 25000 /* TODO!(sqs) */)
                                  .map(({ uri, text }) => {
                                      const diagnostics: sourcegraph.Diagnostic[] = parseDependencies(
                                          text,
                                          globToRegExpDepName(context.packageName)
                                      ).map<sourcegraph.Diagnostic>(({ range, ...dep }) => ({
                                          resource: new URL(uri),
                                          message: `npm dependency '${dep.name}' is deprecated`,
                                          detail: `see campaign [${context.campaignName}](#)`,
                                          range: range,
                                          severity: sourcegraph.DiagnosticSeverity.Warning,
                                          data: JSON.stringify(dep),
                                          tags: [
                                              DEPENDENCY_TAG,
                                              dep.name,
                                              dep.name.replace(/\..*$/, '') /** TODO!(sqs): for lodash */,
                                          ],
                                      }))
                                      return diagnostics
                                  })
                          )
                      )
                  )
              }),
              switchMap(results => results),
              startWith(LOADING)
          )
        : of([])

function globToRegExp(glob: string): RegExp {
    if (glob.endsWith('.*')) {
        return new RegExp(`"${glob.slice(0, -2)}`)
    }
    return new RegExp(`"${glob}"`)
}

function globToRegExpDepName(glob: string): RegExp {
    if (glob.endsWith('.*')) {
        return new RegExp(`^${glob.slice(0, -2)}(\\.|$)`)
    }
    return new RegExp(`^${glob}$`)
}

function globToRegExpObjectProperty(glob: string): RegExp {
    if (glob.endsWith('.*')) {
        return new RegExp(`"${glob.slice(0, -2)}[^"]*":`, 'g')
    }
    return new RegExp(`"${glob}":`, 'g')
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(isProviderDiagnostic)
            if (!diag) {
                return of([])
            }
            return of([
                {
                    title: `Remove dependency from package.json (further edits required)`,
                    edit: computeRemoveDependencyEdit(diag, doc),
                    computeEdit: { title: 'Remove', command: REMOVE_COMMAND },
                    diagnostics: [diag],
                },
            ])
        },
    }
}

interface Dependency {
    name: string
}

/**
 * Parses and returns all dependencies from a package.json file.
 */
function parseDependencies(text: string, packageName?: RegExp): (Dependency & { range: sourcegraph.Range })[] {
    try {
        const data = JSON.parse(text)
        const depNames = sortedUniq([
            ...Object.keys(data.dependencies || {}),
            ...Object.keys(data.devDependencies || {}),
            ...Object.keys(data.peerDependencies || {}),
        ])
        return depNames
            .filter(name => packageName.test(name))
            .map(name => ({ name, range: findDependencyMatchRange(text, name) }))
            .filter(propertyIsDefined('range'))
    } catch (err) {
        // TODO!(sqs): better error handling
        console.error('Error parsing package.json:', err)
        return []
    }
}

function findDependencyMatchRange(text: string, depName: string): sourcegraph.Range | null {
    for (const [i, line] of text.split('\n').entries()) {
        const pat = globToRegExpObjectProperty(depName)
        const match = pat.exec(line)
        if (match) {
            return new sourcegraph.Range(i, match.index, i, match.index + match[0].length)
        }
    }
    return null
    // throw new Error(`dependency ${depName} not found in package.json`)
}

function isProviderDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return diag.tags && diag.tags.includes(DEPENDENCY_TAG)
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): Dependency {
    return JSON.parse(diag.data!)
}

function computeRemoveDependencyEdit(
    diag: sourcegraph.Diagnostic,
    doc: sourcegraph.TextDocument,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    const dep = getDiagnosticData(diag)
    const range = findDependencyMatchRange(doc.text, dep.name)
    // TODO!(sqs): assumes dependency key-value is all on one line and only appears once
    edit.delete(
        new URL(doc.uri),
        new sourcegraph.Range(
            range.start.with({ character: 0 }),
            range.end.with({ line: range.end.line + 1, character: 0 })
        )
    )
    return edit
}

import * as sourcegraph from 'sourcegraph'
import { flatten, sortedUniq } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from } from 'rxjs'
import { map, switchMap, startWith, toArray, filter } from 'rxjs/operators'
import { settingsObservable, memoizedFindTextInFiles } from './util'
import { REPO_INCLUDE } from './misc'

const REMOVE_COMMAND = 'packageJsonDependency.remove'

interface Settings {}

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('packageJsonDependency', {
            provideDiagnostics: _scope =>
                diagnostics.pipe(
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

const diagnostics: Observable<sourcegraph.Diagnostic[] | typeof LOADING> = from(sourcegraph.workspace.rootChanges).pipe(
    startWith(void 0),
    map(() => sourcegraph.workspace.roots),
    switchMap(async roots => {
        if (roots.length > 0) {
            return of<sourcegraph.Diagnostic[]>([]) // TODO!(sqs): dont run in comparison mode
        }

        const results = flatten(
            await from(
                memoizedFindTextInFiles(
                    { pattern: '[Dd]ependencies"', type: 'regexp' },
                    {
                        repositories: {
                            includes: [],
                            excludes: ['hackathon'],
                            type: 'regexp',
                        },
                        files: {
                            includes: ['(^|/)package.json$'],
                            type: 'regexp',
                        },
                        maxResults: 3,
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
                    docs.map(({ uri, text }) => {
                        const diagnostics: sourcegraph.Diagnostic[] = parseDependencies(text).map<
                            sourcegraph.Diagnostic
                        >(({ range, ...dep }) => ({
                            resource: new URL(uri),
                            message: `npm dependency '${dep.name}'`,
                            range: range,
                            severity: sourcegraph.DiagnosticSeverity.Warning,
                            data: JSON.stringify(dep),
                            tags: [DEPENDENCY_TAG],
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

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(isDependencyRulesDiagnostic)
            if (!diag) {
                return of([])
            }
            return from(settingsObservable<Settings>()).pipe(
                map(() => [
                    {
                        title: `Remove dependency from package.json (further edits required)`,
                        edit: computeRemoveDependencyEdit(diag, doc),
                        diagnostics: [diag],
                    },
                ])
            )
        },
    }
}

interface Dependency {
    name: string
}

/**
 * Parses and returns all dependencies from a package.json file.
 */
function parseDependencies(text: string): (Dependency & { range: sourcegraph.Range })[] {
    try {
        const data = JSON.parse(text)
        const depNames = sortedUniq([
            ...Object.keys(data.dependencies || {}),
            ...Object.keys(data.devDependencies || {}),
            ...Object.keys(data.peerDependencies || {}),
        ])
        return depNames.map(name => ({ name, range: findDependencyMatchRange(text, name) }))
    } catch (err) {
        // TODO!(sqs): better error handling
        console.error('Error parsing package.json:', err)
        return []
    }
}

function findDependencyMatchRange(text: string, depName: string): sourcegraph.Range {
    for (const [i, line] of text.split('\n').entries()) {
        const pat = new RegExp(`"${depName}"`, 'g')
        const match = pat.exec(line)
        if (match) {
            return new sourcegraph.Range(i, match.index, i, match.index + match[0].length)
        }
    }
    throw new Error(`dependency ${depName} not found in package.json`)
}

function isDependencyRulesDiagnostic(diag: sourcegraph.Diagnostic): boolean {
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

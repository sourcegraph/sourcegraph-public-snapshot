import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../shared/src/util/types'
import { combineLatestOrDefault } from '../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { flatten } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from } from 'rxjs'
import { map, switchMap, startWith, first, toArray } from 'rxjs/operators'
import { queryGraphQL } from './util'
import * as GQL from '../../../../shared/src/graphql/schema'
import { OTHER_CODE_ACTIONS, MAX_RESULTS, REPO_INCLUDE } from './misc'

export function registerImportStar(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    return subscriptions
}

const mods = [{ binding: 'React', module: 'react' }, { binding: 'H', module: 'history' }]

const CODE_IMPORT_STAR = 'IMPORT_STAR'

function startDiagnostics(): Unsubscribable {
    const subscriptions = new Subscription()

    const diagnosticsCollection = sourcegraph.languages.createDiagnosticCollection('demo0')
    subscriptions.add(diagnosticsCollection)
    subscriptions.add(
        from(sourcegraph.workspace.rootChanges)
            .pipe(
                startWith(void 0),
                map(() => sourcegraph.workspace.roots),
                switchMap(async () => {
                    const results = flatten(
                        await from(
                            sourcegraph.search.findTextInFiles(
                                { pattern: 'import \\* as (React|H)', type: 'regexp' },
                                {
                                    repositories: {
                                        includes: [REPO_INCLUDE],
                                        type: 'regexp',
                                    },
                                    files: {
                                        // includes: ['^(web/src/org|browser/src/libs/phabricator)/.*\\.tsx?$'],
                                        excludes: ['page'],
                                        type: 'regexp',
                                    },
                                    maxResults: MAX_RESULTS,
                                }
                            )
                        )
                            .pipe(toArray())
                            .toPromise()
                    )
                    return combineLatestOrDefault(
                        results.map(async ({ uri }) => {
                            const { text } = await sourcegraph.workspace.openTextDocument(new URL(uri))
                            const diagnostics: sourcegraph.Diagnostic[] = flatten(
                                mods.map(({ binding, module }) =>
                                    findMatchRanges(text, binding, module).map(
                                        range =>
                                            ({
                                                message:
                                                    'Unnecessary `import * as ...` of module that has default export',
                                                range,
                                                severity: sourcegraph.DiagnosticSeverity.Information,
                                                code: CODE_IMPORT_STAR + ':' + JSON.stringify({ binding, module }),
                                            } as sourcegraph.Diagnostic)
                                    )
                                )
                            )
                            return [new URL(uri), diagnostics] as [URL, sourcegraph.Diagnostic[]]
                        })
                    ).pipe(map(items => items.filter(isDefined)))
                }),
                switchMap(results => results)
            )
            .subscribe(entries => {
                diagnosticsCollection.set(entries)
            })
    )

    return diagnosticsCollection
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: async (doc, _rangeOrSelection, context): Promise<sourcegraph.CodeAction[]> => {
            const diag = context.diagnostics.find(isImportStarDiagnostic)
            if (!diag) {
                return []
            }

            const fixAllAction = await computeFixAllAction()

            return [
                {
                    title: 'Convert to named import',
                    edit: computeFixEdit(diag, doc),
                    diagnostics: [diag],
                },
                fixAllAction.edit && Array.from(fixAllAction.edit.textEdits()).length > 1
                    ? {
                          title: 'Fix all unnecessary import-stars',
                          ...fixAllAction,
                      }
                    : null,
                {
                    title: 'Ignore',
                    edit: computeIgnoreEdit(diag, doc),
                    diagnostics: [diag],
                },
                {
                    title: `View npm package: ${getDiagnosticData(diag).module}`,
                    command: { title: '', command: 'TODO!(sqs)' },
                },
                ...OTHER_CODE_ACTIONS,
            ].filter(isDefined)
        },
    }
}

function isImportStarDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return typeof diag.code === 'string' && diag.code.startsWith(CODE_IMPORT_STAR + ':')
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): { binding: string; module: string } {
    const { binding, module } = JSON.parse((diag.code as string).slice((CODE_IMPORT_STAR + ':').length))
    return { binding, module }
}

function computeFixEdit(
    diag: sourcegraph.Diagnostic,
    doc: sourcegraph.TextDocument,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    const { binding, module } = getDiagnosticData(diag)
    for (const range of findMatchRanges(doc.text, binding, module)) {
        edit.replace(new URL(doc.uri), range, `import ${binding} from '${module}'`)
    }
    return edit
}

async function computeFixAllAction(): Promise<Pick<sourcegraph.CodeAction, 'edit' | 'diagnostics'>> {
    // TODO!(sqs): Make this listen for new diagnostics and include those too, but that might be
    // super inefficient because it's n^2, so maybe an altogether better/different solution is
    // needed.
    const allImportStarDiags = sourcegraph.languages
        .getDiagnostics()
        .map(([uri, diagnostics]) => {
            const matchingDiags = diagnostics.filter(isImportStarDiagnostic)
            return matchingDiags.length > 0
                ? ([uri, matchingDiags] as ReturnType<typeof sourcegraph.languages.getDiagnostics>[0])
                : null
        })
        .filter(isDefined)
    const edit = new sourcegraph.WorkspaceEdit()
    for (const [uri, diags] of allImportStarDiags) {
        const doc = await sourcegraph.workspace.openTextDocument(uri)
        for (const diag of diags) {
            computeFixEdit(diag, doc, edit)
        }
    }
    return { edit, diagnostics: flatten(allImportStarDiags.map(([uri, diagnostics]) => diagnostics)) }
}

function computeIgnoreEdit(
    diag: sourcegraph.Diagnostic,
    doc: sourcegraph.TextDocument,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    const { binding, module } = getDiagnosticData(diag)
    for (const range of findMatchRanges(doc.text, binding, module)) {
        edit.insert(
            new URL(doc.uri),
            range.end,
            ' // sourcegraph:ignore-line React lint https://sourcegraph.example.com/ofYRz6NFzj'
        )
    }
    return edit
}

function findMatchRanges(text: string, binding: string, module: string): sourcegraph.Range[] {
    const ranges: sourcegraph.Range[] = []
    for (const [i, line] of text.split('\n').entries()) {
        const pat = new RegExp(`^import \\* as ${binding} from '${module}'$`, 'g')
        for (let match = pat.exec(line); !!match; match = pat.exec(line)) {
            ranges.push(new sourcegraph.Range(i, match.index, i, match.index + match[0].length))
        }
    }
    return ranges
}

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
            const diag = context.diagnostics.find(
                d => typeof d.code === 'string' && d.code.startsWith(CODE_IMPORT_STAR + ':')
            )
            if (!diag) {
                return []
            }
            const { binding, module } = JSON.parse((diag.code as string).slice((CODE_IMPORT_STAR + ':').length))

            const fixEdits = new sourcegraph.WorkspaceEdit()
            for (const range of findMatchRanges(doc.text, binding, module)) {
                fixEdits.replace(new URL(doc.uri), range, `import ${binding} from '${module}'`)
            }

            const disableRuleEdits = new sourcegraph.WorkspaceEdit()

            for (const range of findMatchRanges(doc.text, binding, module)) {
                disableRuleEdits.insert(
                    new URL(doc.uri),
                    range.end,
                    ' // sourcegraph:ignore-line React lint Hi Aneesh and Charlie https://sourcegraph.example.com/ofYRz6NFzj'
                )
            }

            return [
                {
                    title: 'Convert to named import',
                    edit: fixEdits,
                    diagnostics: flatten(
                        sourcegraph.languages.getDiagnostics().map(([uri, diagnostics]) => diagnostics)
                    ),
                },
                {
                    title: 'Ignore',
                    edit: disableRuleEdits,
                    diagnostics: flatten(
                        sourcegraph.languages.getDiagnostics().map(([uri, diagnostics]) => diagnostics)
                    ),
                },
                {
                    title: `View npm package: ${module}`,
                    command: { title: '', command: 'TODO!(sqs)' },
                },
                ...OTHER_CODE_ACTIONS,
            ]
        },
    }
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

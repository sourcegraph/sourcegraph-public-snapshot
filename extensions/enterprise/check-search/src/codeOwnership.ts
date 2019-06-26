import * as sourcegraph from 'sourcegraph'
import { gql, GraphQLResult, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import { isDefined, propertyIsDefined } from '../../../../shared/src/util/types'
import { combineLatestOrDefault } from '../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { flatten, sortedUniq } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from } from 'rxjs'
import { map, switchMap, startWith, first, toArray } from 'rxjs/operators'
import { queryGraphQL, settingsObservable, memoizedFindTextInFiles } from './util'
import * as GQL from '../../../../shared/src/graphql/schema'
import { OTHER_CODE_ACTIONS, MAX_RESULTS, REPO_INCLUDE } from './misc'
import { parseRepoURI, makeRepoURI } from '../../../../shared/src/util/url'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { parsePatch } from 'diff'

export function registerCodeOwnership(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    return subscriptions
}

interface Settings {}

const CODE_CODE_OWNERSHIP_RULES = 'CODE_OWNERSHIP'

function startDiagnostics(): Unsubscribable {
    const subscriptions = new Subscription()

    const diagnosticsCollection = sourcegraph.languages.createDiagnosticCollection('codeOwnership')
    subscriptions.add(diagnosticsCollection)
    subscriptions.add(
        from(sourcegraph.workspace.rootChanges)
            .pipe(
                startWith(void 0),
                map(() => sourcegraph.workspace.roots.filter(propertyIsDefined('baseUri'))),
                switchMap(async roots => {
                    return combineLatestOrDefault(
                        roots
                            .map(async root => {
                                const base = parseRepoURI(root.baseUri.toString())
                                const head = parseRepoURI(root.uri.toString())
                                const data = dataOrThrowErrors<GQL.IQuery>(
                                    await queryGraphQL({
                                        query: gql`
                                            query ComparisonRawDiff(
                                                $repositoryName: String!
                                                $baseRevSpec: String!
                                                $headRevSpec: String!
                                            ) {
                                                repository(name: $repositoryName) {
                                                    comparison(base: $baseRevSpec, head: $headRevSpec) {
                                                        fileDiffs {
                                                            rawDiff
                                                        }
                                                    }
                                                }
                                            }
                                        `,
                                        vars: {
                                            repositoryName: base.repoName,
                                            baseRevSpec: base.rev || base.commitID,
                                            headRevSpec: head.rev || head.commitID,
                                        },
                                    })
                                )
                                const { rawDiff } = data.repository.comparison.fileDiffs
                                const fileDiffs = parsePatch(rawDiff)
                                return Promise.all(
                                    fileDiffs.map(async fileDiff => {
                                        const uri = new URL(makeRepoURI({ ...head, filePath: fileDiff.newFileName }))
                                        const doc = await sourcegraph.workspace.openTextDocument(uri)
                                        const lines = doc.text.split('\n')
                                        const diagnostics = fileDiff.hunks
                                            .map<sourcegraph.Diagnostic | undefined>(hunk => {
                                                const CONTEXT_LINES = 2
                                                const line = hunk.newStart + CONTEXT_LINES
                                                const m = lines[line].match(/\S/)
                                                return {
                                                    message: `MY DIAGNOSTIC IN CHANGESET`,
                                                    range:
                                                        m &&
                                                        doc.getWordRangeAtPosition(
                                                            new sourcegraph.Position(line, m.index)
                                                        ),
                                                    severity: sourcegraph.DiagnosticSeverity.Hint,
                                                    code:
                                                        CODE_CODE_OWNERSHIP_RULES +
                                                        ':' +
                                                        JSON.stringify({ codeOwner: 'alice' } as DiagnosticData),
                                                }
                                            })
                                            .filter(isDefined)
                                        return [uri, diagnostics] as [URL, sourcegraph.Diagnostic[]]
                                    })
                                ).catch(() => [])
                            })
                            .filter(isDefined)
                    )
                }),
                switchMap(results => results),
                map(results => flatten(results))
            )
            .subscribe(entries => {
                diagnosticsCollection.set(entries || [])
            })
    )

    return diagnosticsCollection
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.CodeAction[]> => {
            const diag = context.diagnostics.find(isCodeOwnershipDiagnostic)
            if (!diag) {
                return of([])
            }
            const data = getDiagnosticData(diag)
            return from(settingsObservable<Settings>()).pipe(
                map(settings => {
                    return [
                        {
                            title: `Ask ${data.codeOwner} for review on ${parseRepoURI(doc.uri).filePath}`,
                            command: { title: '', command: 'TODO!(sqs)' },
                        },
                        ...OTHER_CODE_ACTIONS,
                    ].filter(isDefined)
                })
            )
        },
    }
}

interface DiagnosticData {
    codeOwner: string
}

function isCodeOwnershipDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return typeof diag.code === 'string' && diag.code.startsWith(CODE_CODE_OWNERSHIP_RULES + ':')
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): DiagnosticData {
    return JSON.parse((diag.code as string).slice((CODE_CODE_OWNERSHIP_RULES + ':').length))
}

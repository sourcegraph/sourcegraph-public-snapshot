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
import { parsePatch, ParsedDiff, Hunk } from 'diff'

export function registerCodeOwnership(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    return subscriptions
}

interface Settings {}

const TAG_CODE_OWNERSHIP_RULES = 'TAG_OWNERSHIP'

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
                            .map(async (root, rootI) => {
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
                                    fileDiffs.map(async (fileDiff, fileDiffI) => {
                                        const uri = new URL(makeRepoURI({ ...head, filePath: fileDiff.newFileName }))
                                        const doc = await sourcegraph.workspace.openTextDocument(uri)
                                        const diagnostics = [
                                            ...markEnsureAuthz(doc),
                                            ...markSecurityReviewRequired(doc, fileDiff),
                                            ...(rootI === 0 && fileDiffI === 0 ? [suggestChangelogEntry()] : []),
                                            ...(rootI === 0 && fileDiffI === 0 ? updateDependents() : []),
                                            ...markTODO(doc, fileDiff),
                                        ].filter(isDefined)
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
                try {
                    diagnosticsCollection.set(entries || [])
                } catch (err) {
                    console.error(err) // ".for is not iterable" TODO!(sqs)
                }
            })
    )

    return diagnosticsCollection
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(isCodeOwnershipDiagnostic)
            if (!diag) {
                return of([])
            }
            const data = getDiagnosticData(diag)
            return from(settingsObservable<Settings>()).pipe(
                map(settings => {
                    return [
                        ...(data.securityReviewRequired
                            ? [
                                  {
                                      title: `Waiting on review from @${data.codeOwner || 'evan'} (notified 2h ago)`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                              ]
                            : data.authzChecks
                            ? [
                                  {
                                      title: `Self-certify`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `Request review of authz checks from @${data.codeOwner ||
                                          'ziyang71'} (code owner)`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `View internal dev docs on authz`,
                                      command: {
                                          title: '',
                                          command: 'open',
                                          arguments: ['https://docs.sourcegraph.com/dev/architecture#frontend-code'],
                                      },
                                      diagnostics: [diag],
                                  },
                              ]
                            : data.suggestChangelogEntry
                            ? [
                                  {
                                      title: `Not needed (change is minor)`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `Not needed (other reason)`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `Ask @christina (PM) if this needs a changelog entry`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `Add changelog entry...`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `View changelog (for inspiration)`,
                                      command: {
                                          title: '',
                                          command: 'open',
                                          arguments: [
                                              'https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md',
                                          ],
                                      },
                                      diagnostics: [diag],
                                  },
                              ]
                            : data.updateDependents
                            ? [
                                  {
                                      title: `Include in this changeset (default)`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `Create changesets after merging`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `Don't update users`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                                  {
                                      title: `View 17 users`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                              ]
                            : [
                                  {
                                      title: `File as issue and add issue URL`,
                                      command: { title: '', command: 'TODO!(sqs)' },
                                      diagnostics: [diag],
                                  },
                              ]),
                    ].filter(isDefined)
                })
            )
        },
    }
}

interface DiagnosticData {
    codeOwner: string
    securityReviewRequired?: boolean
    authzChecks?: boolean
    suggestChangelogEntry?: boolean
    updateDependents?: boolean
}

function isCodeOwnershipDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return diag.tags && diag.tags.includes(TAG_CODE_OWNERSHIP_RULES)
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): DiagnosticData {
    return JSON.parse(diag.data!)
}

function markSecurityReviewRequired(doc: sourcegraph.TextDocument, fileDiff: ParsedDiff): sourcegraph.Diagnostic[] {
    const ranges = flatten(fileDiff.hunks.map(hunk => findHunkMatchRanges(hunk, /error/g)))
    return ranges.map<sourcegraph.Diagnostic | undefined>(range => {
        return {
            message: `Independent security review required (PCI-compliant code depends on this file)`,
            range,
            severity: sourcegraph.DiagnosticSeverity.Error,
            data: JSON.stringify({
                securityReviewRequired: true,
                codeOwner: range.start.line % 2 === 0 ? 'tsenart' : 'keegan',
            } as DiagnosticData),
            tags: [TAG_CODE_OWNERSHIP_RULES],
        }
    })
}

function markEnsureAuthz(doc: sourcegraph.TextDocument): sourcegraph.Diagnostic[] {
    if (doc.text.includes('SECURITY') || doc.text.includes('security') || doc.text.includes('ListByRepo')) {
        return [
            {
                message: `Ensure changes preserve/add appropriate GraphQL API authorization and security checks`,
                severity: sourcegraph.DiagnosticSeverity.Error,
                data: JSON.stringify({ authzChecks: true } as DiagnosticData),
                tags: [TAG_CODE_OWNERSHIP_RULES],
            },
        ]
    }
    return []
}

function markTODO(doc: sourcegraph.TextDocument, fileDiff: ParsedDiff): sourcegraph.Diagnostic[] {
    const ranges = flatten(fileDiff.hunks.map(hunk => findHunkMatchRanges(hunk, /TODO/g)))
    return ranges.map<sourcegraph.Diagnostic | undefined>(range => {
        return {
            message: `Remove TODO from code before merging`,
            range,
            severity: sourcegraph.DiagnosticSeverity.Hint,
            data: JSON.stringify({ codeOwner: 'alice' } as DiagnosticData),
            tags: [TAG_CODE_OWNERSHIP_RULES],
        }
    })
}

function suggestChangelogEntry(): sourcegraph.Diagnostic {
    return {
        message: `Add changelog entry? (Looks like you changed something user-facing.)`,
        severity: sourcegraph.DiagnosticSeverity.Hint,
        data: JSON.stringify({ suggestChangelogEntry: true } as DiagnosticData),
        tags: [TAG_CODE_OWNERSHIP_RULES],
    }
}

function updateDependents(): sourcegraph.Diagnostic[] {
    return [
        {
            message: `Update users of the changed code?`,
            severity: sourcegraph.DiagnosticSeverity.Information,
            data: JSON.stringify({ updateDependents: true } as DiagnosticData),
            tags: [TAG_CODE_OWNERSHIP_RULES],
        },
    ]
}

function findHunkMatchRanges(hunk: Hunk, pattern: RegExp): sourcegraph.Range[] {
    const ranges: sourcegraph.Range[] = []
    let lineNumber = hunk.newStart - 2
    for (const [i, line] of hunk.lines.entries()) {
        if (line[0] !== '-') {
            lineNumber++
        }
        if (line[0] !== '+') {
            continue // only look at added lines
        }
        pattern.lastIndex = 0
        const match = pattern.exec(line.slice(1))
        if (match) {
            ranges.push(new sourcegraph.Range(lineNumber, match.index, lineNumber, match.index + match[0].length))
        }
    }
    return ranges
}

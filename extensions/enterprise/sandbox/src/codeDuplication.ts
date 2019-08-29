import * as sourcegraph from 'sourcegraph'
import { flatten, sortedUniq, sortBy } from 'lodash'
import { Subscription, Observable, of, Unsubscribable, from, combineLatest } from 'rxjs'
import { map, switchMap, startWith, first, toArray, filter } from 'rxjs/operators'
import { queryGraphQL, settingsObservable, memoizedFindTextInFiles } from './util'
import * as GQL from '../../../../shared/src/graphql/schema'
import { OTHER_CODE_ACTIONS, MAX_RESULTS, REPO_INCLUDE } from './misc'
import { JSCPD, IClone, IOptions, MATCH_SOURCE_EVENT, CLONE_FOUND_EVENT, SOURCE_SKIPPED_EVENT, END_EVENT } from 'jscpd'
import { IStoreManagerOptions } from 'jscpd/build/interfaces/store/store-manager-options.interface'
import { ITokenLocation } from 'jscpd/build/interfaces/token/token-location.interface'
import { StoresManager } from 'jscpd/build/stores/stores-manager'

const FILE_ISSUE_COMMAND = 'codeDuplication.fileIssue'
const IGNORE_COMMAND = 'codeDuplication.ignore'

interface Settings {}

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(startDiagnostics())
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(FILE_ISSUE_COMMAND, () => new sourcegraph.WorkspaceEdit())
    )
    subscriptions.add(sourcegraph.commands.registerActionEditCommand(IGNORE_COMMAND, ignoreCommandCallback))
    return subscriptions
}

const LOADING: 'loading' = 'loading'

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
                    { pattern: '', type: 'regexp' },
                    {
                        repositories: {
                            // includes: [],
                            excludes: ['about|/sourcegraph/'], // exclude forks
                            type: 'regexp',
                        },
                        files: {
                            includes: ['\\.(go|[jt]sx?)$'], // TODO!(sqs): typescript only
                            excludes: ['\\.pb\\.go$'], // exclude protobuf-generated files
                            type: 'regexp',
                        },
                        maxResults: 75, //MAX_RESULTS,
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
            switchMap(async () => {
                StoresManager.close()
                StoresManager.flush()
                const cpd = new JSCPD({})
                const clones: IClone[] = []
                for (const doc of docs) {
                    clones.push(...(await cpd.detect(doc.text, { id: doc.uri, format: jscpdFormat(doc) })))
                }
                const diagnostics: sourcegraph.Diagnostic[] = clones.map(c => {
                    const numLines = c.duplicationA.end.line - c.duplicationA.start.line
                    return {
                        resource: new URL(c.duplicationA.sourceId),
                        range: duplicationRange(c.duplicationA),
                        message: `Duplicated code (${numLines} line${numLines !== 1 ? 's' : ''})`,
                        source: 'codeDuplication',
                        severity: sourcegraph.DiagnosticSeverity.Information,
                        relatedInformation: [
                            {
                                location: new sourcegraph.Location(
                                    new URL(c.duplicationB.sourceId),
                                    duplicationRange(c.duplicationB)
                                ),
                                message: 'Duplicated here',
                            },
                        ],
                        data: JSON.stringify(c),
                        tags: [c.format],
                    } as sourcegraph.Diagnostic
                })
                return diagnostics
            })
        )
    }),
    switchMap(results => results),
    startWith(LOADING)
)

const jscpdFormat = (doc: sourcegraph.TextDocument): string => {
    switch (doc.languageId) {
        case 'typescriptreact':
            return 'tsx'
        case 'javascriptreact':
            return 'jsx'
    }
    return doc.languageId
}

function startDiagnostics(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('codeDuplication', {
            provideDiagnostics: _scope =>
                diagnostics.pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    return subscriptions
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (doc, _rangeOrSelection, context): Observable<sourcegraph.Action[]> => {
            const diag = context.diagnostics.find(isCodeDuplicationDiagnostic)
            if (!diag) {
                return of([])
            }
            return from(settingsObservable<Settings>()).pipe(
                map(settings => {
                    const actions: sourcegraph.Action[] = [
                        {
                            title: 'File on code owners (@alice @bob)',
                            command: { title: 'File issue', command: FILE_ISSUE_COMMAND },
                            diagnostics: [diag],
                        },
                        {
                            title: 'Ignore',
                            computeEdit: { title: 'Ignore', command: IGNORE_COMMAND },
                            diagnostics: [diag],
                        },
                    ]
                    return actions
                })
            )
        },
    }
}

function duplicationRange({ start, end }: { start: ITokenLocation; end: ITokenLocation }): sourcegraph.Range {
    return new sourcegraph.Range(
        new sourcegraph.Position(start.line, start.column || 0),
        new sourcegraph.Position(end.line, end.column || 0)
    )
}

function isCodeDuplicationDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    // TODO!(sqs)
    return true
}

function createWorkspaceEditForIgnore(
    doc: sourcegraph.TextDocument,
    diagnostic: sourcegraph.Diagnostic,
    edit = new sourcegraph.WorkspaceEdit()
): sourcegraph.WorkspaceEdit {
    // TODO!(sqs): get indent of previous line - in vscode this is inserted on the client
    // automatically, check out how they do it because that seems neat
    // (https://sourcegraph.com/github.com/microsoft/vscode-tslint@30d1a7ae25b0331466f1a54b4f7d23d60fa2da30/-/blob/tslint-server/src/tslintServer.ts#L618)

    const range = diagnostic.range // TODO!(sqs): get other duplication instance too

    const startIndent = doc.text.slice(doc.offsetAt(range.start.with(undefined, 0))).match(/[ \t]*/)
    edit.insert(
        new URL(doc.uri),
        range.start.with(undefined, 0),
        `${startIndent ? startIndent[0] : ''}// jscpd:ignore-start\n`
    )

    const endIndent = doc.text.slice(doc.offsetAt(range.end.with(undefined, 0))).match(/[ \t]*/)
    edit.insert(
        new URL(doc.uri),
        doc.positionAt(1 + doc.offsetAt(range.end)),
        `${endIndent ? endIndent[0] : ''}// jscpd:ignore-end\n`
    )
    return edit
}

async function ignoreCommandCallback(diagnostic: sourcegraph.Diagnostic): Promise<sourcegraph.WorkspaceEdit> {
    const doc = await sourcegraph.workspace.openTextDocument(diagnostic.resource)
    return createWorkspaceEditForIgnore(doc, diagnostic)
}

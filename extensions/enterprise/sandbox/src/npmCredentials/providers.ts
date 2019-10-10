import { flatten } from 'lodash'
import { from, Observable, Subscription, Unsubscribable } from 'rxjs'
import { filter, map, startWith, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { isDefined } from '../../../../../shared/src/util/types'
import { scanForCredentials } from './scanner'

const NPM_CREDENTIALS_FIX_COMMAND = 'npmCredentials.fix'

export interface NPMCredentialsCampaignContext {
    filters?: string
}

const LOADING = 'loading' as const

export function register(): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        sourcegraph.workspace.registerDiagnosticProvider('npmCredentials', {
            provideDiagnostics: (_scope, context) =>
                provideDiagnostics((context as any) as NPMCredentialsCampaignContext).pipe(
                    filter((diagnostics): diagnostics is sourcegraph.Diagnostic[] => diagnostics !== LOADING)
                ),
        })
    )
    subscriptions.add(sourcegraph.languages.registerCodeActionProvider(['*'], createCodeActionProvider()))
    subscriptions.add(
        sourcegraph.commands.registerActionEditCommand(NPM_CREDENTIALS_FIX_COMMAND, diagnostic => {
            if (!diagnostic) {
                return new sourcegraph.WorkspaceEdit()
            }
            return computeFixEdit(diagnostic)
        })
    )
    return subscriptions
}

const TAG = 'type:npmCredentials'

interface DiagnosticData {}

function provideDiagnostics({
    filters,
}: NPMCredentialsCampaignContext): Observable<sourcegraph.Diagnostic[] | typeof LOADING> {
    return from(sourcegraph.workspace.rootChanges).pipe(
        startWith(undefined),
        map(() => sourcegraph.workspace.roots),
        switchMap(async roots => {
            if (roots.length > 0) {
                return [] as sourcegraph.Diagnostic[] // TODO!(sqs): dont run in comparison mode
            }

            const results = await scanForCredentials({ filters })
            return flatten(
                (await Promise.all(
                    results.map(async result => {
                        const doc = await sourcegraph.workspace.openTextDocument(new URL(result.uri))

                        // Skip minified files.
                        if (doc.text!.split('\n').some(line => line.length > 300)) {
                            return null
                        }

                        const range = findTokenRange(doc.text!)
                        if (range) {
                            const diagnostic: sourcegraph.Diagnostic = {
                                resource: new URL(result.uri),
                                message: 'npm credential must not be committed to source control',
                                detail:
                                    'revoke with `curl -uadmin:password -XPOST "http://example.com/artifactory/api/security/token/revoke" -d "token=TOKEN"`',
                                range,
                                severity: sourcegraph.DiagnosticSeverity.Error,
                                // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                                data: JSON.stringify({} as DiagnosticData),
                                tags: [TAG, 'checkbox'],
                            }
                            return [diagnostic]
                        }
                        return null
                    })
                )).filter(isDefined)
            )
        }),
        startWith(LOADING)
    )
}

function createCodeActionProvider(): sourcegraph.CodeActionProvider {
    return {
        provideCodeActions: (_doc, _rangeOrSelection, context): sourcegraph.Action[] => {
            const diag = context.diagnostics.find(isProviderDiagnostic)
            if (!diag) {
                return []
            }
            return [
                {
                    title: 'Remove npm credentials',
                    edit: computeFixEdit(diag),
                    computeEdit: { title: 'Remove npm credentials', command: NPM_CREDENTIALS_FIX_COMMAND },
                    diagnostics: [diag],
                },
            ]
        },
    }
}

function findTokenRange(text: string, str: string): sourcegraph.Range | null {
    for (const [i, line] of text.split('\n').entries()) {
        const j = line.indexOf(str)
        if (j !== -1) {
            return new sourcegraph.Range(i, j, i, j + str.length)
        }
    }
    return null
}

function isProviderDiagnostic(diag: sourcegraph.Diagnostic): boolean {
    return !!diag.tags && diag.tags.includes(TAG)
}

function getDiagnosticData(diag: sourcegraph.Diagnostic): DiagnosticData {
    return JSON.parse(diag.data!)
}

function computeFixEdit(diag: sourcegraph.Diagnostic): sourcegraph.WorkspaceEdit {
    const edit = new sourcegraph.WorkspaceEdit()
    edit.replace(diag.resource, diag.range, 'REDACTED')
    return edit
}

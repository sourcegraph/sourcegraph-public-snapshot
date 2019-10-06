import { Range } from '@sourcegraph/extension-api-classes'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { DiagnosticWithType } from '../../../../shared/src/api/client/services/diagnosticService'
import { match } from '../../../../shared/src/api/client/types/textDocument'
import { Action, fromAction } from '../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { getModeFromPath } from '../../../../shared/src/languages'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'

export function diagnosticQueryMatcher(
    query: sourcegraph.DiagnosticQuery
): (diagnostic: DiagnosticWithType) => boolean {
    return diagnostic =>
        diagnostic.type === query.type &&
        (!query.document ||
            match(query.document, {
                uri: diagnostic.resource.toString(),
                languageId: getModeFromPath(diagnostic.resource.pathname),
            })) &&
        (!query.range || query.range.isEqual(diagnostic.range)) &&
        (query.tag !== undefined ? query.tag.every(tag => diagnostic.tags && diagnostic.tags.includes(tag)) : true) &&
        (query.message === undefined || diagnostic.message.toLowerCase().includes(query.message.toLowerCase()))
}

// TODO!(sqs): this assumes a diag provider never has 2 diagnostics on the same range and resource
export const diagnosticID = (diagnostic: DiagnosticWithType): string =>
    `${diagnostic.type}:${diagnostic.resource.toString()}:${diagnostic.range ? diagnostic.range.start.line : '-'}:${
        diagnostic.range ? diagnostic.range.start.character : '-'
    }`

export const diagnosticQueryForSingleDiagnostic = (diagnostic: DiagnosticWithType): sourcegraph.DiagnosticQuery => ({
    type: diagnostic.type,
    document: [{ pattern: diagnostic.resource.toString() }],
    range: diagnostic.range,
    tag: diagnostic.tags,
})

export const getCodeActions = memoizeObservable(
    ({
        diagnostic,
        extensionsController,
    }: { diagnostic: DiagnosticWithType } & ExtensionsControllerProps): Observable<Action[]> =>
        from(
            extensionsController.services.codeActions.getCodeActions({
                textDocument: {
                    uri: diagnostic.resource.toString(),
                },
                range: Range.fromPlain(diagnostic.range),
                context: { diagnostics: [diagnostic] },
            })
        ).pipe(
            map(codeActions => codeActions || []),
            map(actions => actions.map(fromAction))
        ),
    ({ diagnostic }) => diagnosticID(diagnostic)
)

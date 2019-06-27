import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import { Range, Selection } from '@sourcegraph/extension-api-classes'
import * as clientType from '@sourcegraph/extension-api-types'
import { Unsubscribable } from 'rxjs'
import {
    CodeActionProvider,
    CompletionItemProvider,
    DefinitionProvider,
    DocumentSelector,
    HoverProvider,
    Location,
    LocationProvider,
    ReferenceProvider,
    CodeAction,
    ChecklistProvider,
} from 'sourcegraph'
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures'
import { CodeActionsParams } from '../../client/services/codeActions'
import { ReferenceParams, TextDocumentPositionParams } from '../../protocol'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'
import { ExtDocuments } from './documents'
import { fromCodeAction, fromHover, fromLocation, toPosition } from './types'
import { WorkspaceEdit } from '../../types/workspaceEdit'

/** @internal */
export class ExtChecklist {
    constructor(private proxy: ProxyResult<ClientChecklistAPI>) {}

    public registerChecklistProvider(type: string, provider: ChecklistProvider): Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientChecklistAPI['$registerChecklistProvider']>[2]
        > = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideLocations(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        return syncSubscription(this.proxy.$registerLocationProvider(idStr, selector, proxyValue(providerFunction)))
    }
}

function toLocations(result: Location[] | Location | null | undefined): clientType.Location[] {
    return result ? (Array.isArray(result) ? result : [result]).map(location => fromLocation(location)) : []
}

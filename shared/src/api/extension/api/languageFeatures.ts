/* eslint-disable no-sync */
import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as clientType from '@sourcegraph/extension-api-types'
import { Unsubscribable } from 'rxjs'
import {
    CompletionItemProvider,
    DefinitionProvider,
    DocumentSelector,
    HoverProvider,
    Location,
    LocationProvider,
    ReferenceProvider,
} from 'sourcegraph'
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures'
import { ReferenceParams, TextDocumentPositionParams } from '../../protocol'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'
import { ExtDocuments } from './documents'
import { fromHover, fromLocation, toPosition } from './types'

/** @internal */
export class ExtLanguageFeatures {
    constructor(private proxy: ProxyResult<ClientLanguageFeaturesAPI>, private documents: ExtDocuments) {}

    public registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable {
        const providerFunction: ProxyInput<Parameters<
            ClientLanguageFeaturesAPI['$registerHoverProvider']
        >[1]> = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideHover(await this.documents.getSync(textDocument.uri), toPosition(position)),
                hover => (hover ? fromHover(hover) : hover)
            )
        )
        return syncSubscription(this.proxy.$registerHoverProvider(selector, providerFunction))
    }

    public registerDefinitionProvider(selector: DocumentSelector, provider: DefinitionProvider): Unsubscribable {
        const providerFunction: ProxyInput<Parameters<
            ClientLanguageFeaturesAPI['$registerDefinitionProvider']
        >[1]> = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideDefinition(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        return syncSubscription(this.proxy.$registerDefinitionProvider(selector, providerFunction))
    }

    public registerReferenceProvider(selector: DocumentSelector, provider: ReferenceProvider): Unsubscribable {
        const providerFunction: ProxyInput<Parameters<
            ClientLanguageFeaturesAPI['$registerReferenceProvider']
        >[1]> = proxyValue(async ({ textDocument, position, context }: ReferenceParams) =>
            toProxyableSubscribable(
                provider.provideReferences(
                    await this.documents.getSync(textDocument.uri),
                    toPosition(position),
                    context
                ),
                toLocations
            )
        )
        return syncSubscription(this.proxy.$registerReferenceProvider(selector, providerFunction))
    }

    public registerLocationProvider(
        idStr: string,
        selector: DocumentSelector,
        provider: LocationProvider
    ): Unsubscribable {
        const providerFunction: ProxyInput<Parameters<
            ClientLanguageFeaturesAPI['$registerLocationProvider']
        >[2]> = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideLocations(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        return syncSubscription(this.proxy.$registerLocationProvider(idStr, selector, proxyValue(providerFunction)))
    }

    public registerCompletionItemProvider(
        selector: DocumentSelector,
        provider: CompletionItemProvider
    ): Unsubscribable {
        const providerFunction: ProxyInput<Parameters<
            ClientLanguageFeaturesAPI['$registerCompletionItemProvider']
        >[1]> = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideCompletionItems(await this.documents.getSync(textDocument.uri), toPosition(position)),
                items => items
            )
        )
        return syncSubscription(this.proxy.$registerCompletionItemProvider(selector, providerFunction))
    }
}

function toLocations(result: Location[] | Location | null | undefined): clientType.Location[] {
    return result ? (Array.isArray(result) ? result : [result]).map(location => fromLocation(location)) : []
}

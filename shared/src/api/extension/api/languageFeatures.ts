import * as clientType from '@sourcegraph/extension-api-types'
import { ProxyInput, ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { Subscription, Unsubscribable } from 'rxjs'
import {
    DefinitionProvider,
    DocumentSelector,
    HoverProvider,
    ImplementationProvider,
    Location,
    LocationProvider,
    ReferenceProvider,
    TypeDefinitionProvider,
} from 'sourcegraph'
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures'
import { ReferenceParams, TextDocumentPositionParams } from '../../protocol'
import { toProxyableSubscribable } from './common'
import { ExtDocuments } from './documents'
import { fromHover, fromLocation, toPosition } from './types'

/** @internal */
export interface ExtLanguageFeaturesAPI {}

/** @internal */
export class ExtLanguageFeatures implements ExtLanguageFeaturesAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    constructor(private proxy: ProxyResult<ClientLanguageFeaturesAPI>, private documents: ExtDocuments) {}

    public registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientLanguageFeaturesAPI['$registerHoverProvider']>[1]
        > = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideHover(await this.documents.getSync(textDocument.uri), toPosition(position)),
                hover => (hover ? fromHover(hover) : hover)
            )
        )
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy.$registerHoverProvider(selector, providerFunction).then(s => subscription.add(s))
        return subscription
    }

    public registerDefinitionProvider(selector: DocumentSelector, provider: DefinitionProvider): Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientLanguageFeaturesAPI['$registerDefinitionProvider']>[1]
        > = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideDefinition(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy.$registerDefinitionProvider(selector, providerFunction).then(s => subscription.add(s))
        return subscription
    }

    public registerTypeDefinitionProvider(
        selector: DocumentSelector,
        provider: TypeDefinitionProvider
    ): Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientLanguageFeaturesAPI['$registerTypeDefinitionProvider']>[1]
        > = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideTypeDefinition(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy.$registerTypeDefinitionProvider(selector, providerFunction).then(s => subscription.add(s))
        return subscription
    }

    public registerImplementationProvider(
        selector: DocumentSelector,
        provider: ImplementationProvider
    ): Unsubscribable {
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy
            .$registerImplementationProvider(
                selector,
                proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
                    toProxyableSubscribable(
                        provider.provideImplementation(
                            await this.documents.getSync(textDocument.uri),
                            toPosition(position)
                        ),
                        toLocations
                    )
                )
            )
            .then(s => subscription.add(s))
        return subscription
    }

    public registerReferenceProvider(selector: DocumentSelector, provider: ReferenceProvider): Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientLanguageFeaturesAPI['$registerReferenceProvider']>[1]
        > = proxyValue(async ({ textDocument, position, context }: ReferenceParams) =>
            toProxyableSubscribable(
                provider.provideReferences(
                    await this.documents.getSync(textDocument.uri),
                    toPosition(position),
                    context
                ),
                toLocations
            )
        )
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy.$registerReferenceProvider(selector, providerFunction).then(s => subscription.add(s))
        return subscription
    }

    public registerLocationProvider(
        idStr: string,
        selector: DocumentSelector,
        provider: LocationProvider
    ): Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientLanguageFeaturesAPI['$registerLocationProvider']>[2]
        > = proxyValue(async ({ textDocument, position }: TextDocumentPositionParams) =>
            toProxyableSubscribable(
                provider.provideLocations(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy
            .$registerLocationProvider(idStr, selector, proxyValue(providerFunction))
            .then(s => subscription.add(s))
        return subscription
    }
}

function toLocations(result: Location[] | Location | null | undefined): clientType.Location[] {
    return result ? (Array.isArray(result) ? result : [result]).map(location => fromLocation(location)) : []
}

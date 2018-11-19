import { Observable, Unsubscribable } from 'rxjs'
import {
    Definition,
    DefinitionProvider,
    DocumentSelector,
    Hover,
    HoverProvider,
    ImplementationProvider,
    Location,
    ReferenceContext,
    ReferenceProvider,
    Subscribable,
    TypeDefinitionProvider,
} from 'sourcegraph'
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures'
import * as plain from '../../protocol/plainTypes'
import { ProviderMap, toProviderResultObservable } from './common'
import { ExtDocuments } from './documents'
import { fromHover, fromLocation, toPosition } from './types'

/** @internal */
export interface ExtLanguageFeaturesAPI {
    $observeHover(id: number, resource: string, position: plain.Position): Observable<plain.Hover | null | undefined>
    $observeDefinition(id: number, resource: string, position: plain.Position): Observable<plain.Definition | undefined>
    $observeTypeDefinition(
        id: number,
        resource: string,
        position: plain.Position
    ): Observable<plain.Definition | undefined>
    $observeImplementation(
        id: number,
        resource: string,
        position: plain.Position
    ): Observable<plain.Definition | undefined>
    $observeReferences(
        id: number,
        resource: string,
        position: plain.Position,
        context: ReferenceContext
    ): Observable<plain.Location[] | null | undefined>
}

/** @internal */
export class ExtLanguageFeatures implements ExtLanguageFeaturesAPI {
    private registrations = new ProviderMap<
        HoverProvider | DefinitionProvider | TypeDefinitionProvider | ImplementationProvider | ReferenceProvider
    >(id => this.proxy.$unregister(id))

    constructor(private proxy: ClientLanguageFeaturesAPI, private documents: ExtDocuments) {}

    public $observeHover(
        id: number,
        resource: string,
        position: plain.Position
    ): Observable<plain.Hover | null | undefined> {
        const provider = this.registrations.get<HoverProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<Hover | undefined | null | Subscribable<Hover | undefined | null>>(document =>
                    provider.provideHover(document, toPosition(position))
                ),
            hover => (hover ? fromHover(hover) : hover)
        )
    }

    public registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerHoverProvider(id, selector)
        return subscription
    }

    public $observeDefinition(
        id: number,
        resource: string,
        position: plain.Position
    ): Observable<plain.Definition | undefined> {
        const provider = this.registrations.get<DefinitionProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<Definition | undefined | Subscribable<Definition | undefined>>(document =>
                    provider.provideDefinition(document, toPosition(position))
                ),
            toDefinition
        )
    }

    public registerDefinitionProvider(selector: DocumentSelector, provider: DefinitionProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerDefinitionProvider(id, selector)
        return subscription
    }

    public $observeTypeDefinition(
        id: number,
        resource: string,
        position: plain.Position
    ): Observable<plain.Definition | null | undefined> {
        const provider = this.registrations.get<TypeDefinitionProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<Definition | undefined | Subscribable<Definition | undefined>>(document =>
                    provider.provideTypeDefinition(document, toPosition(position))
                ),
            toDefinition
        )
    }

    public registerTypeDefinitionProvider(
        selector: DocumentSelector,
        provider: TypeDefinitionProvider
    ): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerTypeDefinitionProvider(id, selector)
        return subscription
    }

    public $observeImplementation(
        id: number,
        resource: string,
        position: plain.Position
    ): Observable<plain.Definition | undefined> {
        const provider = this.registrations.get<ImplementationProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<Definition | undefined | Subscribable<Definition | undefined>>(document =>
                    provider.provideImplementation(document, toPosition(position))
                ),
            toDefinition
        )
    }

    public registerImplementationProvider(
        selector: DocumentSelector,
        provider: ImplementationProvider
    ): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerImplementationProvider(id, selector)
        return subscription
    }

    public $observeReferences(
        id: number,
        resource: string,
        position: plain.Position,
        context: ReferenceContext
    ): Observable<plain.Location[] | null | undefined> {
        const provider = this.registrations.get<ReferenceProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<Location[] | null | undefined | Subscribable<Location[] | null | undefined>>(document =>
                    provider.provideReferences(document, toPosition(position), context)
                ),
            toLocations
        )
    }

    public registerReferenceProvider(selector: DocumentSelector, provider: ReferenceProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerReferenceProvider(id, selector)
        return subscription
    }
}

function toLocations(result: Location[] | null | undefined): plain.Location[] | null | undefined {
    return result ? result.map(location => fromLocation(location)) : result
}

function toDefinition(result: Location[] | Location | null | undefined): plain.Definition | undefined {
    return result
        ? Array.isArray(result)
            ? result.map(location => fromLocation(location))
            : fromLocation(result)
        : result
}

import { Unsubscribable } from 'rxjs'
import {
    DefinitionProvider,
    DocumentSelector,
    HoverProvider,
    ImplementationProvider,
    Location,
    ReferenceContext,
    ReferenceProvider,
    TypeDefinitionProvider,
} from 'sourcegraph'
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures'
import * as plain from '../../protocol/plainTypes'
import { ProviderMap } from './common'
import { ExtDocuments } from './documents'
import { fromHover, fromLocation, toPosition } from './types'

/** @internal */
export interface ExtLanguageFeaturesAPI {
    $provideHover(id: number, resource: string, position: plain.Position): Promise<plain.Hover | null | undefined>
    $provideDefinition(id: number, resource: string, position: plain.Position): Promise<plain.Definition | undefined>
    $provideTypeDefinition(
        id: number,
        resource: string,
        position: plain.Position
    ): Promise<plain.Definition | undefined>
    $provideImplementation(
        id: number,
        resource: string,
        position: plain.Position
    ): Promise<plain.Definition | undefined>
    $provideReferences(
        id: number,
        resource: string,
        position: plain.Position,
        context: ReferenceContext
    ): Promise<plain.Location[] | null | undefined>
}

/** @internal */
export class ExtLanguageFeatures implements ExtLanguageFeaturesAPI {
    private registrations = new ProviderMap<
        HoverProvider | DefinitionProvider | TypeDefinitionProvider | ImplementationProvider | ReferenceProvider
    >(id => this.proxy.$unregister(id))

    constructor(private proxy: ClientLanguageFeaturesAPI, private documents: ExtDocuments) {}

    public async $provideHover(
        id: number,
        resource: string,
        position: plain.Position
    ): Promise<plain.Hover | null | undefined> {
        const provider = this.registrations.get<HoverProvider>(id)
        return Promise.resolve(
            provider.provideHover(await this.documents.getSync(resource), toPosition(position))
        ).then(result => (result ? fromHover(result) : result))
    }

    public registerHoverProvider(selector: DocumentSelector, provider: HoverProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerHoverProvider(id, selector)
        return subscription
    }

    public async $provideDefinition(
        id: number,
        resource: string,
        position: plain.Position
    ): Promise<plain.Definition | null | undefined> {
        const provider = this.registrations.get<DefinitionProvider>(id)
        return Promise.resolve(
            provider.provideDefinition(await this.documents.getSync(resource), toPosition(position))
        ).then(toDefinition)
    }

    public registerDefinitionProvider(selector: DocumentSelector, provider: DefinitionProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerDefinitionProvider(id, selector)
        return subscription
    }

    public async $provideTypeDefinition(
        id: number,
        resource: string,
        position: plain.Position
    ): Promise<plain.Definition | null | undefined> {
        const provider = this.registrations.get<TypeDefinitionProvider>(id)
        return Promise.resolve(
            provider.provideTypeDefinition(await this.documents.getSync(resource), toPosition(position))
        ).then(toDefinition)
    }

    public registerTypeDefinitionProvider(
        selector: DocumentSelector,
        provider: TypeDefinitionProvider
    ): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerTypeDefinitionProvider(id, selector)
        return subscription
    }

    public async $provideImplementation(
        id: number,
        resource: string,
        position: plain.Position
    ): Promise<plain.Definition | undefined> {
        const provider = this.registrations.get<ImplementationProvider>(id)
        return Promise.resolve(
            provider.provideImplementation(await this.documents.getSync(resource), toPosition(position))
        ).then(toDefinition)
    }

    public registerImplementationProvider(
        selector: DocumentSelector,
        provider: ImplementationProvider
    ): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerImplementationProvider(id, selector)
        return subscription
    }

    public async $provideReferences(
        id: number,
        resource: string,
        position: plain.Position,
        context: ReferenceContext
    ): Promise<plain.Location[] | null | undefined> {
        const provider = this.registrations.get<ReferenceProvider>(id)
        return Promise.resolve(
            provider.provideReferences(await this.documents.getSync(resource), toPosition(position), context)
        ).then(toLocations)
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

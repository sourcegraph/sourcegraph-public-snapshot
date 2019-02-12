import * as clientType from '@sourcegraph/extension-api-types'
import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { from, Observable, observable, PartialObserver, Subscription, Unsubscribable } from 'rxjs'
import {
    DefinitionProvider,
    DocumentSelector,
    Hover,
    HoverProvider,
    ImplementationProvider,
    Location,
    LocationProvider,
    ProviderResult,
    ReferenceContext,
    ReferenceProvider,
    Subscribable,
    TypeDefinitionProvider,
} from 'sourcegraph'
import { ClientLanguageFeaturesAPI } from '../../client/api/languageFeatures'
import { isSubscribable } from '../../util'
import { ProviderMap, toProviderResultObservable } from './common'
import { ExtDocuments } from './documents'
import { fromHover, fromLocation, toPosition } from './types'

/** @internal */
export interface ExtLanguageFeaturesAPI {
    $observeHover(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Hover | null | undefined>
    $observeDefinition(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined>
    $observeTypeDefinition(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined>
    $observeImplementation(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined>
    $observeReferences(
        id: number,
        resource: string,
        position: clientType.Position,
        context: ReferenceContext
    ): Observable<clientType.Location[] | null | undefined>
    $observeLocations(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined>
}

interface SubscribableNoOverloads<T> {
    subscribe(
        ...observer:
            | [PartialObserver<T> | undefined]
            | [
                  ((value: T) => void) | undefined | null,
                  ((error: any) => void) | undefined | null,
                  (() => void) | undefined | null
              ]
    ): Subscription
}

const createRemoteObservable = <T>(proxy: ProxyResult<SubscribableNoOverloads<T>>): Observable<T> =>
    from(({
        [observable](): Subscribable<T> {
            return this
        },
        subscribe(observer: PartialObserver<T>): Subscription {
            const subscription = new Subscription()
            proxy.subscribe(comlink.proxyValue(observer)).then(s => {
                subscription.add(s)
            })
            return subscription
        },
    } as any) as Subscribable<T>)

class ProxySubscribable<T> implements ProxyValue, Subscribable<T> {
    public readonly [proxyValueSymbol] = true

    constructor(private subscribable: Subscribable<T>) {}

    public subscribe(observer?: PartialObserver<T>): Unsubscribable
    public subscribe(
        next?: (value: T) => void,
        error?: (error: any) => void,
        complete?: () => void
    ): Unsubscribable & ProxyValue
    public subscribe(...args: any[]): Unsubscribable & ProxyValue {
        return proxyValue(this.subscribable.subscribe(...args))
    }
}

const proxyProviderFunction = <P extends any[], T>(
    fn: (...args: P) => ProviderResult<T>
): ((...args: P) => T | undefined | null | Promise<T | undefined | null> | ProxySubscribable<T | null | undefined>) &
    ProxyValue =>
    proxyValue((...args: P) => {
        const result = fn(...args)
        if (isSubscribable(result)) {
            return new ProxySubscribable(result)
        }
        return result
    })

/** @internal */
export class ExtLanguageFeatures implements ExtLanguageFeaturesAPI, Unsubscribable, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private registrations = new ProviderMap<
        | HoverProvider
        | DefinitionProvider
        | TypeDefinitionProvider
        | ImplementationProvider
        | ReferenceProvider
        | LocationProvider
    >(id => this.proxy.$unregister(id))

    constructor(private proxy: ProxyResult<ClientLanguageFeaturesAPI>, private documents: ExtDocuments) {}

    public $observeHover(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Hover | null | undefined> {
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
        const subscription = new Subscription()
        // tslint:disable-next-line:no-floating-promises
        this.proxy
            .$registerHoverProvider(selector, proxyProviderFunction(provider.provideHover.bind(provider)))
            .then(s => subscription.add(s))
        return subscription
    }

    public $observeDefinition(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined> {
        const provider = this.registrations.get<DefinitionProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<
                    Location | Location[] | null | undefined | Subscribable<Location | Location[] | null | undefined>
                >(document => provider.provideDefinition(document, toPosition(position))),
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
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined> {
        const provider = this.registrations.get<TypeDefinitionProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<
                    Location | Location[] | null | undefined | Subscribable<Location | Location[] | null | undefined>
                >(document => provider.provideTypeDefinition(document, toPosition(position))),
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
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined> {
        const provider = this.registrations.get<ImplementationProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<
                    Location | Location[] | null | undefined | Subscribable<Location | Location[] | null | undefined>
                >(document => provider.provideImplementation(document, toPosition(position))),
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
        position: clientType.Position,
        context: ReferenceContext
    ): Observable<clientType.Location[] | null | undefined> {
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

    public $observeLocations(
        id: number,
        resource: string,
        position: clientType.Position
    ): Observable<clientType.Location[] | null | undefined> {
        const provider = this.registrations.get<LocationProvider>(id)
        return toProviderResultObservable(
            this.documents
                .getSync(resource)
                .then<Location[] | null | undefined | Subscribable<Location[] | null | undefined>>(document =>
                    provider.provideLocations(document, toPosition(position))
                ),
            toLocations
        )
    }

    public registerLocationProvider(
        idStr: string,
        selector: DocumentSelector,
        provider: LocationProvider
    ): Unsubscribable {
        /**
         * {@link idStr} is the `id` parameter to {@link sourcegraph.languages.registerLocationProvider} that
         * identifies the provider and is chosen by the extension. {@link id} is an internal implementation detail:
         * the numeric registry ID used to identify this provider solely between the client and extension host.
         */
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerLocationProvider(id, idStr, selector)
        return subscription
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}

function toLocations(result: Location[] | null | undefined): clientType.Location[] | null | undefined {
    return result ? result.map(location => fromLocation(location)) : result
}

function toDefinition(result: Location[] | Location | null | undefined): clientType.Location[] | null | undefined {
    return result ? (Array.isArray(result) ? result : [result]).map(location => fromLocation(location)) : result
}

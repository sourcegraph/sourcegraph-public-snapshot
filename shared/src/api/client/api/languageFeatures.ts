import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Hover, Location } from '@sourcegraph/extension-api-types'
import { map } from 'rxjs/operators'
import { CodeAction, CompletionList, DocumentSelector, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { toCodeAction } from '../../extension/api/types'
import { ReferenceParams, TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { CodeActionsParams, ProvideCodeActionsSignature } from '../services/codeActions'
import { ProvideCompletionItemSignature } from '../services/completion'
import { ProvideTextDocumentHoverSignature } from '../services/hover'
import { TextDocumentLocationProviderIDRegistry, TextDocumentLocationProviderRegistry } from '../services/location'
import { FeatureProviderRegistry } from '../services/registry'
import { wrapRemoteObservable } from './common'

/** @internal */
export interface ClientLanguageFeaturesAPI extends ProxyValue {
    $registerHoverProvider(
        selector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Hover | null | undefined>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue
    $registerDefinitionProvider(
        selector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue
    $registerReferenceProvider(
        selector: DocumentSelector,
        providerFunction: ProxyResult<((params: ReferenceParams) => ProxySubscribable<Location[]>) & ProxyValue>
    ): Unsubscribable & ProxyValue

    /**
     * @param idStr The `id` argument in the extension's {@link sourcegraph.languages.registerLocationProvider}
     * call.
     */
    $registerLocationProvider(
        idStr: string,
        selector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue

    $registerCompletionItemProvider(
        selector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<CompletionList | null | undefined>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue

    $registerCodeActionProvider(
        selector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: CodeActionsParams) => ProxySubscribable<CodeAction[] | null | undefined>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

/** @internal */
export class ClientLanguageFeatures implements ClientLanguageFeaturesAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    constructor(
        private hoverRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentHoverSignature
        >,
        private definitionRegistry: TextDocumentLocationProviderRegistry,
        private referencesRegistry: TextDocumentLocationProviderRegistry<ReferenceParams>,
        private locationRegistry: TextDocumentLocationProviderIDRegistry,
        private completionItemsRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideCompletionItemSignature
        >,
        private codeActionsRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideCodeActionsSignature
        >
    ) {}

    public $registerHoverProvider(
        documentSelector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Hover | null | undefined>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.hoverRegistry.registerProvider({ documentSelector }, params =>
                wrapRemoteObservable(providerFunction(params))
            )
        )
    }

    public $registerDefinitionProvider(
        documentSelector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.definitionRegistry.registerProvider({ documentSelector }, params =>
                wrapRemoteObservable(providerFunction(params))
            )
        )
    }

    public $registerReferenceProvider(
        documentSelector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.referencesRegistry.registerProvider({ documentSelector }, params =>
                wrapRemoteObservable(providerFunction(params))
            )
        )
    }

    public $registerLocationProvider(
        id: string,
        documentSelector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.locationRegistry.registerProvider({ id, documentSelector }, params =>
                wrapRemoteObservable(providerFunction(params))
            )
        )
    }

    public $registerCompletionItemProvider(
        documentSelector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: TextDocumentPositionParams) => ProxySubscribable<CompletionList | null | undefined>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.completionItemsRegistry.registerProvider({ documentSelector }, params =>
                wrapRemoteObservable(providerFunction(params))
            )
        )
    }

    public $registerCodeActionProvider(
        documentSelector: DocumentSelector,
        providerFunction: ProxyResult<
            ((params: CodeActionsParams) => ProxySubscribable<CodeAction[] | null | undefined>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.codeActionsRegistry.registerProvider({ documentSelector }, params =>
                wrapRemoteObservable(providerFunction({ ...params, range: (params.range as any).toJSON() })).pipe(
                    map(codeActions =>
                        codeActions ? codeActions.map(codeAction => toCodeAction(codeAction)) : codeActions
                    )
                )
            )
        )
    }
}

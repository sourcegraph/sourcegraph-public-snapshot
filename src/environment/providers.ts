import { BehaviorSubject, combineLatest, from, Observable, TeardownLogic } from 'rxjs'
import { catchError, first, map, switchMap } from 'rxjs/operators'
import { Hover } from 'vscode-languageserver-types'
import { CommandRegistry } from '../client/features/commands'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../protocol'
import { TextDocumentDecoration, TextDocumentDecorationsParams } from '../protocol/decorations'
import { HoverMerged } from '../types/hover'
import { compact, flatten } from '../util'

interface Entry<O extends TextDocumentRegistrationOptions, P> {
    registrationOptions: O
    provider: P
}

export abstract class TextDocumentFeatureProviderRegistry<O extends TextDocumentRegistrationOptions, P> {
    private entries = new BehaviorSubject<Entry<O, P>[]>([])

    public registerProvider(registrationOptions: O, provider: P): TeardownLogic {
        const entry: Entry<O, P> = { registrationOptions, provider }
        this.entries.next([...this.entries.value, entry])
        return () => {
            this.entries.next(this.entries.value.filter(e => e !== entry))
        }
    }

    /** All providers, emitted whenever the set of registered providers changed. */
    public readonly providers: Observable<P[]> = this.entries.pipe(
        map(providers => providers.map(({ provider }) => provider))
    )

    /**
     * The current set of providers. Used by callers that do not need to react to providers being registered or
     * unregistered.
     */
    public readonly providersSnapshot = this.providers.pipe(first())
}

export type ProvideTextDocumentHoverSignature = (params: TextDocumentPositionParams) => Promise<Hover | null>

class TextDocumentHoverProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentHoverSignature
> {
    public getHover(params: TextDocumentPositionParams): Observable<HoverMerged | null> {
        return this.providersSnapshot
            .pipe(
                switchMap(providers =>
                    combineLatest(
                        providers.map(provider =>
                            from(provider(params)).pipe(
                                catchError(error => {
                                    console.error(error)
                                    return [null]
                                })
                            )
                        )
                    )
                )
            )
            .pipe(map(HoverMerged.from))
    }
}

export type ProvideTextDocumentDecorationsSignature = (
    params: TextDocumentDecorationsParams
) => Observable<TextDocumentDecoration[] | null>

class TextDocumentDecorationsProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentDecorationsSignature
> {
    public getDecorations(params: TextDocumentDecorationsParams): Observable<TextDocumentDecoration[] | null> {
        return this.providers
            .pipe(
                switchMap(providers =>
                    combineLatest(
                        providers.map(provider =>
                            provider(params).pipe(
                                map(results => (results === null ? [] : compact(results))),
                                catchError(error => {
                                    console.error(error)
                                    return [[]]
                                })
                            )
                        )
                    )
                )
            )
            .pipe(map(results => flatten(results)))
    }
}

export class NoopProviderRegistry extends TextDocumentFeatureProviderRegistry<any, any> {}

/** Registries is a container for all provider registries. */
export class Registries {
    public readonly commands = new CommandRegistry()
    public readonly textDocumentHover = new TextDocumentHoverProviderRegistry()
    public readonly textDocumentDecorations = new TextDocumentDecorationsProviderRegistry()
}

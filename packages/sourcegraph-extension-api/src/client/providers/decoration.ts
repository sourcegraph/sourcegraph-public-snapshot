import { combineLatest, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import {
    DecorationAttachmentRenderOptions,
    ThemableDecorationAttachmentStyle,
    ThemableDecorationStyle,
} from 'sourcegraph'
import { TextDocumentIdentifier } from '../../client/types/textDocument'
import { TextDocumentDecoration } from '../../protocol/plainTypes'
import { FeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

export type ProvideTextDocumentDecorationSignature = (
    textDocument: TextDocumentIdentifier
) => Observable<TextDocumentDecoration[] | null>

/** Provides text document decorations from all extensions. */
export class TextDocumentDecorationProviderRegistry extends FeatureProviderRegistry<
    undefined,
    ProvideTextDocumentDecorationSignature
> {
    public getDecorations(params: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null> {
        return getDecorations(this.providers, params)
    }
}

/**
 * Returns an observable that emits all decorations whenever any of the last-emitted set of providers emits
 * decorations.
 *
 * Most callers should use TextDocumentDecorationProviderRegistry, which uses the registered decoration providers.
 */
export function getDecorations(
    providers: Observable<ProvideTextDocumentDecorationSignature[]>,
    params: TextDocumentIdentifier
): Observable<TextDocumentDecoration[] | null> {
    return providers
        .pipe(
            switchMap(providers => {
                if (providers.length === 0) {
                    return [null]
                }
                return combineLatest(providers.map(provider => provider(params)))
            })
        )
        .pipe(map(flattenAndCompact))
}

/**
 * Resolves the actual styles to use for the attachment based on the current theme.
 */
export function decorationStyleForTheme(
    attachment: TextDocumentDecoration,
    isLightTheme: boolean
): ThemableDecorationStyle {
    const overrides = isLightTheme ? attachment.light : attachment.dark
    if (!overrides) {
        return attachment
    }
    // Discard non-ThemableDecorationStyle properties so they aren't included in result.
    const { range, isWholeLine, after, light, dark, ...base } = attachment
    return { ...base, ...overrides }
}

/**
 * Resolves the actual styles to use for the attachment based on the current theme.
 */
export function decorationAttachmentStyleForTheme(
    attachment: DecorationAttachmentRenderOptions,
    isLightTheme: boolean
): ThemableDecorationAttachmentStyle {
    const overrides = isLightTheme ? attachment.light : attachment.dark
    if (!overrides) {
        return attachment
    }
    // Discard non-ThemableDecorationAttachmentStyle properties so they aren't included in result.
    const { contentText, hoverMessage, linkURL, light, dark, ...base } = attachment
    return { ...base, ...overrides }
}

import { Observable } from 'rxjs';
import { DecorationAttachmentRenderOptions, ThemableDecorationAttachmentStyle, ThemableDecorationStyle } from 'sourcegraph';
import { TextDocumentIdentifier } from '../../client/types/textDocument';
import { TextDocumentDecoration } from '../../protocol/plainTypes';
import { FeatureProviderRegistry } from './registry';
export declare type ProvideTextDocumentDecorationSignature = (textDocument: TextDocumentIdentifier) => Observable<TextDocumentDecoration[] | null>;
/** Provides text document decorations from all extensions. */
export declare class TextDocumentDecorationProviderRegistry extends FeatureProviderRegistry<undefined, ProvideTextDocumentDecorationSignature> {
    getDecorations(params: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null>;
}
/**
 * Returns an observable that emits all decorations whenever any of the last-emitted set of providers emits
 * decorations.
 *
 * Most callers should use TextDocumentDecorationProviderRegistry, which uses the registered decoration providers.
 */
export declare function getDecorations(providers: Observable<ProvideTextDocumentDecorationSignature[]>, params: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null>;
/**
 * Resolves the actual styles to use for the attachment based on the current theme.
 */
export declare function decorationStyleForTheme(attachment: TextDocumentDecoration, isLightTheme: boolean): ThemableDecorationStyle;
/**
 * Resolves the actual styles to use for the attachment based on the current theme.
 */
export declare function decorationAttachmentStyleForTheme(attachment: DecorationAttachmentRenderOptions, isLightTheme: boolean): ThemableDecorationAttachmentStyle;

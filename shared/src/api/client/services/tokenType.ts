import { TokenType } from 'sourcegraph'
import { TextDocumentPositionParams } from '../../protocol'
import { DocumentFeatureProviderRegistry } from './registry'

/**
 * TODO
 */
export class TextDocumentTokenTypeProviderRegistry extends DocumentFeatureProviderRegistry<
    (params: TextDocumentPositionParams) => Promise<TokenType>
> {}

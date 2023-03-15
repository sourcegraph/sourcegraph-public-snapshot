import { EmbeddingSearchResult } from '../embeddings'

export interface ContextSearchOptions {
    numCodeResults: number
    numMarkdownResults: number
    filterResults?: (result: EmbeddingSearchResult) => boolean
}

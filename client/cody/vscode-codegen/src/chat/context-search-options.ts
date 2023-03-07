import { EmbeddingSearchResult } from '../embeddings-client'

export interface ContextSearchOptions {
	numCodeResults: number
	numMarkdownResults: number
	filterResults?: (result: EmbeddingSearchResult) => boolean
}

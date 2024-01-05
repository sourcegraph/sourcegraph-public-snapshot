import create from 'zustand'

import type { ContentMatch, PathMatch } from '@sourcegraph/shared/src/search/stream'

/**
 * At the moment search result preview panel supports only
 * blob-like type of search results to preview.
 */
export type SearchResultPreview = ContentMatch | PathMatch

export interface SearchResultState {
    previewBlob: SearchResultPreview | null
    setPreviewBlob: (blobInfo: SearchResultPreview) => void
    clearPreview: () => void
}

export const useSearchResultState = create<SearchResultState>((set, get) => ({
    previewBlob: null,
    setPreviewBlob: blobInfo => set({ previewBlob: blobInfo }),
    clearPreview: () => set({ previewBlob: null }),
}))

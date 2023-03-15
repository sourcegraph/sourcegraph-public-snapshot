import { Message, QueryInfo } from '@sourcegraph/cody-common'

import { ActiveTextEditor, ActiveTextEditorSelection, Editor } from '../../editor'
import { EmbeddingSearchResults, Embeddings } from '../../embeddings'
import { IntentDetector } from '../../intent-detector'
import { KeywordContextFetcher } from '../../keyword-context'

export class MockEmbeddingsClient implements Embeddings {
    constructor(private mocks: Partial<Embeddings> = {}) {}

    search(query: string, codeCount: number, markdownCount: number): Promise<EmbeddingSearchResults> {
        return (
            this.mocks.search?.(query, codeCount, markdownCount) ??
            Promise.resolve({ codeResults: [], markdownResults: [] })
        )
    }

    queryNeedsAdditionalContext(query: string): Promise<boolean> {
        return this.mocks.queryNeedsAdditionalContext?.(query) ?? Promise.resolve(false)
    }
}

export class MockIntentDetector implements IntentDetector {
    constructor(private mocks: Partial<IntentDetector> = {}) {}

    detect(text: string): Promise<QueryInfo> {
        return (
            this.mocks.detect?.(text) ??
            Promise.resolve({ needsCodebaseContext: false, needsCurrentFileContext: false })
        )
    }
}

export class MockKeywordContextFetcher implements KeywordContextFetcher {
    constructor(private mocks: Partial<KeywordContextFetcher> = {}) {}

    getContextMessages(query: string): Promise<Message[]> {
        return this.mocks.getContextMessages?.(query) ?? Promise.resolve([])
    }
}

export class MockEditor implements Editor {
    constructor(private mocks: Partial<Editor> = {}) {}

    getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        return this.mocks.getActiveTextEditorSelection?.() ?? null
    }

    getActiveTextEditor(): ActiveTextEditor | null {
        return this.mocks.getActiveTextEditor?.() ?? null
    }

    showQuickPick(labels: string[]): Promise<string | undefined> {
        return this.mocks.showQuickPick?.(labels) ?? Promise.resolve(undefined)
    }

    showWarningMessage(message: string): Promise<void> {
        return this.mocks.showWarningMessage?.(message) ?? Promise.resolve()
    }
}

export const defaultEmbeddingsClient = new MockEmbeddingsClient()

export const defaultIntentDetector = new MockIntentDetector()

export const defaultKeywordContextFetcher = new MockKeywordContextFetcher()

export const defaultEditor = new MockEditor()

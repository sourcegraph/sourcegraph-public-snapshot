import { ActiveTextEditor, ActiveTextEditorSelection, ActiveTextEditorVisibleContent, Editor } from '../../editor'
import { EmbeddingsSearch } from '../../embeddings'
import { IntentDetector } from '../../intent-detector'
import { KeywordContextFetcher, KeywordContextFetcherResult } from '../../keyword-context'
import { EmbeddingsSearchResults } from '../../sourcegraph-api/graphql'

export class MockEmbeddingsClient implements EmbeddingsSearch {
    constructor(private mocks: Partial<EmbeddingsSearch> = {}) {}

    public search(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<EmbeddingsSearchResults | Error> {
        return (
            this.mocks.search?.(query, codeResultsCount, textResultsCount) ??
            Promise.resolve({ codeResults: [], textResults: [] })
        )
    }
}

export class MockIntentDetector implements IntentDetector {
    constructor(private mocks: Partial<IntentDetector> = {}) {}

    public isCodebaseContextRequired(input: string): Promise<boolean | Error> {
        return this.mocks.isCodebaseContextRequired?.(input) ?? Promise.resolve(false)
    }
}

export class MockKeywordContextFetcher implements KeywordContextFetcher {
    constructor(private mocks: Partial<KeywordContextFetcher> = {}) {}

    public getContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]> {
        return this.mocks.getContext?.(query, numResults) ?? Promise.resolve([])
    }
}

export class MockEditor implements Editor {
    constructor(private mocks: Partial<Editor> = {}) {}

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        return this.mocks.getActiveTextEditorSelection?.() ?? null
    }

    public getActiveTextEditor(): ActiveTextEditor | null {
        return this.mocks.getActiveTextEditor?.() ?? null
    }

    public getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null {
        return this.mocks.getActiveTextEditorVisibleContent?.() ?? null
    }

    public showQuickPick(labels: string[]): Promise<string | undefined> {
        return this.mocks.showQuickPick?.(labels) ?? Promise.resolve(undefined)
    }

    public showWarningMessage(message: string): Promise<void> {
        return this.mocks.showWarningMessage?.(message) ?? Promise.resolve()
    }
}

export const defaultEmbeddingsClient = new MockEmbeddingsClient()

export const defaultIntentDetector = new MockIntentDetector()

export const defaultKeywordContextFetcher = new MockKeywordContextFetcher()

export const defaultEditor = new MockEditor()

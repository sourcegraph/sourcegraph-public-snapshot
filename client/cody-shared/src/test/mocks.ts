import { BotResponseMultiplexer } from '../chat/bot-response-multiplexer'
import { RecipeContext } from '../chat/recipes/recipe'
import { CodebaseContext } from '../codebase-context'
import { ActiveTextEditor, ActiveTextEditorSelection, ActiveTextEditorVisibleContent, Editor } from '../editor'
import { EmbeddingsSearch } from '../embeddings'
import { IntentDetector } from '../intent-detector'
import { KeywordContextFetcher, KeywordContextFetcherResult } from '../keyword-context'
import { EmbeddingsSearchResults } from '../sourcegraph-api/graphql'

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

    public isEditorContextRequired(input: string): boolean | Error {
        return this.mocks.isEditorContextRequired?.(input) ?? false
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

    public getWorkspaceRootPath(): string | null {
        return this.mocks.getWorkspaceRootPath?.() ?? null
    }

    public getActiveTextEditorSelection(): ActiveTextEditorSelection | null {
        return this.mocks.getActiveTextEditorSelection?.() ?? null
    }

    public getActiveTextEditorSelectionOrEntireFile(): ActiveTextEditorSelection | null {
        return this.mocks.getActiveTextEditorSelection?.() ?? null
    }

    public getActiveTextEditor(): ActiveTextEditor | null {
        return this.mocks.getActiveTextEditor?.() ?? null
    }

    public getActiveTextEditorVisibleContent(): ActiveTextEditorVisibleContent | null {
        return this.mocks.getActiveTextEditorVisibleContent?.() ?? null
    }

    public replaceSelection(fileName: string, selectedText: string, replacement: string): Promise<void> {
        return this.mocks.replaceSelection?.(fileName, selectedText, replacement) ?? Promise.resolve()
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

export function newRecipeContext(args?: Partial<RecipeContext>): RecipeContext {
    args = args || {}
    return {
        editor: args.editor || defaultEditor,
        intentDetector: args.intentDetector || defaultIntentDetector,
        codebaseContext:
            args.codebaseContext ||
            new CodebaseContext(
                { useContext: 'none' },
                'dummy-codebase',
                defaultEmbeddingsClient,
                defaultKeywordContextFetcher
            ),
        responseMultiplexer: args.responseMultiplexer || new BotResponseMultiplexer(),
    }
}

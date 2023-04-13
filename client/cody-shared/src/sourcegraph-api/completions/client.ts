import {
    Event,
    CompletionParameters,
    CompletionCallbacks,
    CodeCompletionParameters,
    CodeCompletionResponse,
} from './types'

export abstract class SourcegraphCompletionsClient {
    protected completionsEndpoint: string
    protected codeCompletionsEndpoint: string

    constructor(
        instanceUrl: string,
        protected accessToken: string | null,
        protected mode: 'development' | 'production'
    ) {
        this.completionsEndpoint = `${instanceUrl}/.api/completions/stream`
        this.codeCompletionsEndpoint = `${instanceUrl}/.api/completions/code`
    }

    protected sendEvents(events: Event[], cb: CompletionCallbacks): void {
        for (const event of events) {
            switch (event.type) {
                case 'completion':
                    cb.onChange(event.completion)
                    break
                case 'error':
                    cb.onError(event.error)
                    break
                case 'done':
                    cb.onComplete()
                    break
            }
        }
    }

    public abstract stream(params: CompletionParameters, cb: CompletionCallbacks): () => void
    public abstract complete(params: CodeCompletionParameters): Promise<CodeCompletionResponse>
}

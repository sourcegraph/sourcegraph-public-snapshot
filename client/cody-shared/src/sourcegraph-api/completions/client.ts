import { ConfigurationWithAccessToken } from '../../configuration'

import {
    Event,
    CompletionParameters,
    CompletionCallbacks,
    CodeCompletionParameters,
    CodeCompletionResponse,
} from './types'

type Config = Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'debug' | 'customHeaders'>

export abstract class SourcegraphCompletionsClient {
    constructor(protected config: Config) {}

    public onConfigurationChange(newConfig: Config): void {
        this.config = newConfig
    }

    protected get completionsEndpoint(): string {
        return new URL('/.api/completions/stream', this.config.serverEndpoint).href
    }

    protected get codeCompletionsEndpoint(): string {
        return new URL('/.api/completions/code', this.config.serverEndpoint).href
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

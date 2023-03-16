import { CompletionParameters, SourcegraphCompletionsClient } from '../sourcegraph-api/completions'
import { isError } from '../utils'

import { IntentDetector, QueryInfo } from '.'

export class LLMIntentDetector implements IntentDetector {
    constructor(private completions: SourcegraphCompletionsClient) {}

    public async detect(query: string): Promise<QueryInfo | Error> {
        const [needsCodebaseContext, needsCurrentFileContext] = await Promise.all([
            this.getShortAnswer(
                `The user has a code editor open to a current file inside the current codebase. Does the following question from the user require knowledge of other files in the current codebase (not the current file)? Answer ONLY with a single word, "yes" or "no".\n${query}`
            ),
            this.getShortAnswer(
                `The user has a code editor open to a current file inside the current codebase. Does the following question from the user require knowledge of the current file? Answer ONLY with a single word, "yes" or "no".\n${query}`
            ),
        ])

        if (isError(needsCodebaseContext)) {
            return needsCodebaseContext
        }
        if (isError(needsCurrentFileContext)) {
            return needsCurrentFileContext
        }

        return { needsCodebaseContext, needsCurrentFileContext }
    }

    private getShortAnswer(prompt: string): Promise<boolean | Error> {
        return new Promise((resolve, reject) => {
            let answer = ''
            this.completions.stream(
                {
                    messages: [
                        { speaker: 'human', text: prompt },
                        { speaker: 'assistant', text: '' },
                    ],
                    ...SHORT_ANSWER_COMPLETION_PARAMETERS,
                },
                {
                    onChange: text => {
                        answer = text
                    },
                    onComplete: () => resolve(answer.trim().toLocaleLowerCase().startsWith('yes')),
                    onError: error => reject(new Error(error)),
                }
            )
        })
    }
}

const SHORT_ANSWER_COMPLETION_PARAMETERS: Omit<CompletionParameters, 'messages'> = {
    temperature: 0.0,
    maxTokensToSample: 1,
    topK: -1,
    topP: -1,
}

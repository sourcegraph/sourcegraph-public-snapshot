import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

import { IntentDetector } from '.'

const editorRegexps = [/editor/, /(open|current)\s+file/, /current(ly)?\s+open/, /have\s+open/]

export class SourcegraphIntentDetectorClient implements IntentDetector {
    constructor(private client: SourcegraphGraphQLAPIClient) {}

    public isCodebaseContextRequired(input: string): Promise<boolean | Error> {
        return this.client.isContextRequiredForQuery(input)
    }

    public isEditorContextRequired(input: string): boolean | Error {
        const inputLowerCase = input.toLowerCase()
        // If the input matches any of the `editorRegexps` we assume that we have to include
        // the editor context (e.g., currently open file) to the overall message context.
        for (const regexp of editorRegexps) {
            if (inputLowerCase.match(regexp)) {
                return true
            }
        }
        return false
    }
}

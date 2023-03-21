import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

import { IntentDetector } from '.'

export class SourcegraphIntentDetectorClient implements IntentDetector {
    constructor(private client: SourcegraphGraphQLAPIClient) {}

    isCodebaseContextRequired(input: string): Promise<boolean | Error> {
        return this.client.isContextRequiredForQuery(input)
    }
}

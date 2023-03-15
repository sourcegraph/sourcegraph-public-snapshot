import { Message } from '@sourcegraph/cody-common'

export interface KeywordContextFetcher {
    getContextMessages(query: string): Promise<Message[]>
}

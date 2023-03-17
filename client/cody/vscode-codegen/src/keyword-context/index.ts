import { Message } from '../sourcegraph-api'

export interface KeywordContextFetcher {
    getContextMessages(query: string): Promise<Message[]>
}

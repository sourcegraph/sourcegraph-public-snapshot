import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { LLMReranker } from '@sourcegraph/cody-shared/src/codebase-context/rerank'
import { ContextResult } from '@sourcegraph/cody-shared/src/local-context'

import { debug } from './log'
import { TestSupport } from './test-support'

export function getRerankWithLog(
    chatClient: ChatClient
): (query: string, results: ContextResult[]) => Promise<ContextResult[]> {
    if (TestSupport.instance) {
        const reranker = TestSupport.instance.getReranker()
        return (query: string, results: ContextResult[]): Promise<ContextResult[]> => reranker.rerank(query, results)
    }

    const reranker = new LLMReranker(chatClient)
    return async (userQuery: string, results: ContextResult[]): Promise<ContextResult[]> => {
        const start = performance.now()
        const rerankedResults = await reranker.rerank(userQuery, results)
        const duration = performance.now() - start
        debug('Reranker:rerank', JSON.stringify({ duration }))
        return rerankedResults
    }
}

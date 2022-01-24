import { of, Subscription } from 'rxjs'
import { throttleTime } from 'rxjs/operators'
import * as vscode from 'vscode'

import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'

import { ExtensionCoreAPI } from '../contract'
import { VSCEStateMachine } from '../state'

export function createStreamSearch({
    context,
    stateMachine,
    sourcegraphURL,
}: {
    context: vscode.ExtensionContext
    stateMachine: VSCEStateMachine
    sourcegraphURL: string
}): ExtensionCoreAPI['streamSearch'] {
    // Ensure only one search is active at a time
    let previousSearchSubscription: Subscription | null

    context.subscriptions.push({
        dispose: () => {
            previousSearchSubscription?.unsubscribe()
        },
    })

    return function streamSearch(query, options) {
        previousSearchSubscription?.unsubscribe()

        stateMachine.emit({
            type: 'submit_search_query',
            submittedSearchQueryState: {
                queryState: { query },
                searchCaseSensitivity: options.caseSensitive,
                searchPatternType: options.patternType,
            },
        })

        previousSearchSubscription = aggregateStreamingSearch(of(query), {
            ...options,
            sourcegraphURL,
        })
            .pipe(throttleTime(500, undefined, { leading: true, trailing: true }))
            .subscribe(searchResults => {
                stateMachine.emit({ type: 'received_search_results', searchResults })
            })
    }
}

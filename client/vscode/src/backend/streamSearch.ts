import { of, Subscription } from 'rxjs'
import { throttleTime } from 'rxjs/operators'
import * as vscode from 'vscode'

import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'

import { ExtensionCoreAPI } from '../contract'
import { VSCEStateMachine } from '../state'
import { focusSearchPanel } from '../webview/commands'

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
        // Focus search panel if not already focused
        // (in case e.g. user initiates search from search sidebar when panel is hidden).
        focusSearchPanel()

        previousSearchSubscription = aggregateStreamingSearch(
            of(appendContextFilter(query, stateMachine.state.context.selectedSearchContextSpec)),
            {
                ...options,
                sourcegraphURL,
            }
        )
            .pipe(throttleTime(500, undefined, { leading: true, trailing: true }))
            .subscribe(searchResults => {
                if (searchResults.state === 'error') {
                    // Pass only primitive copied values because Error object is not cloneable
                    const { name, message, stack } = searchResults.error
                    searchResults.error = { name, message, stack }
                }

                stateMachine.emit({ type: 'received_search_results', searchResults })
            })
    }
}

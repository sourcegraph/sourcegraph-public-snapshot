import { from, Observable, of } from 'rxjs'
import { catchError } from 'rxjs/operators'
import * as vscode from 'vscode'

import { GraphQLResult } from '@sourcegraph/http-client'
import { getAvailableSearchContextSpecOrDefault } from '@sourcegraph/search'

import { LocalStorageService, SELECTED_SEARCH_CONTEXT_SPEC_KEY } from '../settings/LocalStorageService'
import { VSCEStateMachine } from '../state'

import { requestGraphQLFromVSCode } from './requestGraphQl'

// Returns an Observable so webviews can easily block rendering on init.
export function initializeSearchContexts({
    localStorageService,
    stateMachine,
    context,
}: {
    localStorageService: LocalStorageService
    stateMachine: VSCEStateMachine
    context: vscode.ExtensionContext
}): void {
    const initialSearchContextSpec = localStorageService.getValue(SELECTED_SEARCH_CONTEXT_SPEC_KEY)

    const defaultSpec = 'global'

    const subscription = getAvailableSearchContextSpecOrDefault({
        spec: initialSearchContextSpec || defaultSpec,
        defaultSpec,
        platformContext: {
            requestGraphQL: ({ request, variables }) =>
                from(requestGraphQLFromVSCode(request, variables)) as Observable<GraphQLResult<any>>,
        },
    })
        .pipe(
            catchError(error => {
                console.error('Error validating search context spec:', error)
                return of(defaultSpec)
            })
        )
        .subscribe(availableSearchContextSpecOrDefault => {
            stateMachine.emit({ type: 'set_selected_search_context_spec', spec: availableSearchContextSpecOrDefault })
        })

    context.subscriptions.push({
        dispose: () => subscription.unsubscribe(),
    })
}

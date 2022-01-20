import { cloneDeep } from 'lodash'
import { BehaviorSubject, Observable } from 'rxjs'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

// State management in the Sourcegraph VS Code extension
// -----
// This extension runs code in 4 (and counting) different execution contexts.
// Coordinating state between these contexts is a difficult task.
// So, instead of managing shared state in each context, we maintain
// one state machine in the "Core" context (see './extension.ts' for architecure diagram).
// All contexts listen for state updates and emit events on which the state
// machine may transition.
// For example:
// - Commands from VS Code extension core
// - The first submitted search in a session will cause the state machine
//    to transition from the `search-home` state to the `search-results` state.
//    This new state will be reflected in both the search sidebar and search panel UIs
//

// We represent a hierarchical state machine in a "flat" manner to reduce code complexity
// and because our state machine is simple enough to not necessitate bringing in a library.
// So,
//   ┌───►home
//   │
// - search
//   │
//   └───►results
// - remote-browsing
// - idle
// - context-invalidated
// becomes:
// - [search-home, search-results, remote-browsing, idle, context-invalidated]

// Example user flow state transitions:
// - User clicks on Sourcegraph logo in VS Code sidebar.
// - Extension activates with initial state of `search-home`
// - User submits search -> state === `search-results`
// - User clicks on a search result, which opens a file -> state === `remote-browsing`
// - User copies some code, then focuses an editor for a local file -> state === `idle`

export interface VSCEStateMachine {
    state: VSCEState
    /**
     * Returns an Observable that emits the current state and
     * on subsequent state updates.
     */
    observeState: () => Observable<VSCEState>
    emit: (event: VSCEEvent) => void
}
export type VSCEState = SearchHomeState | SearchResultsState | RemoteBrowsingState | IdleState | ContextInvalidatedState

interface SearchHomeState {
    status: 'search-home'
    context: CommonContext & {}
}

interface SearchResultsState {
    status: 'search-results'
    context: CommonContext & {}
}

interface RemoteBrowsingState {
    status: 'remote-browsing'
    context: CommonContext & {}
}

interface IdleState {
    status: 'idle'
    context: CommonContext & {}
}

interface ContextInvalidatedState {
    status: 'context-invalidated'
    context: CommonContext & {}
}

interface CommonContext {
    authenticatedUser: AuthenticatedUser | null
}

const INITIAL_STATE: VSCEState = { status: 'search-home', context: { authenticatedUser: null } }

// Temporary placeholder events. We will replace these with the actual events as we implement the webviews.

export type VSCEEvent = AccessTokenValidationEvent | SearchEvent | { type: 'sourcegraph_url_change' }

type AccessTokenValidationEvent =
    | { type: 'validate_access_token'; accessToken: string }
    | { type: 'access_token_validation_failure' }

type SearchEvent = { type: 'set_query_state' } | { type: 'submit_search_query' }

export function createVSCEStateMachine(): VSCEStateMachine {
    const states = new BehaviorSubject<VSCEState>(INITIAL_STATE)

    function reducer(state: VSCEState, event: VSCEEvent): VSCEState {
        if (event.type === 'sourcegraph_url_change') {
            return { status: 'context-invalidated', context: INITIAL_STATE.context }
        }

        switch (state.status) {
            case 'search-home':
                switch (event.type) {
                    case 'submit_search_query':
                        return {
                            ...state,
                            status: 'search-results',
                        }
                }
                return state

            case 'search-results':
                return state

            case 'remote-browsing':
                return state

            case 'idle':
                return state

            case 'context-invalidated':
                return state
        }
    }

    return {
        get state() {
            return cloneDeep(states.value)
        },
        observeState: () => states.asObservable(),
        emit: event => {
            const nextState = reducer(states.value, event)
            states.next(nextState)
        },
    }
}

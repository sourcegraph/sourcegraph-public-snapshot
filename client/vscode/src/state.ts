import { cloneDeep } from 'lodash'
import { BehaviorSubject, Observable } from 'rxjs'

export interface VSCEStateMachine {
    state: VSCEState
    observeState: () => Observable<VSCEState>
    emit: (event: VSCEEvent) => void
}

// Sourcegraph VS Code extension states:
// - extension not activated: this state is implicit in our codebase; simply when this code isn't run.
// - extension activated: state transitions here on activation events (e.g. onCommand:sourcegraph.search or onView:sourcegraph.extensionHost)
//   - searching (context/data: search query, search context, etc.)
//     - search home (only when search panel is freshly opened. once a search has been executed on this panel instance, it's going to be in search results state)
//        - auth'ed (proper search sidebar)
//        - unauth'ed (sign up CTA)
//     - search results
//   - remote file browsing (potential actions from here: going to def, finding refs, cloning repo locally)
//   - idle/not engaged (e.g. editing local file, search panel is not focused)
//   - context invalidated (Sourcegraph URL or access token changed)

export interface VSCEState {
    status: 'search-home' | 'search-results' | 'remote-browsing' | 'idle' | 'context-invalidated'
}

// TODO common context between sub-states (e.g. search query?)

export type VSCEEvent = AccessTokenValidationEvent | SearchEvent | { type: 'sourcegraph_url_change' }

type AccessTokenValidationEvent =
    | { type: 'validate_access_token'; accessToken: string }
    | { type: 'access_token_validation_failure' }

type SearchEvent = { type: 'submit_search_query' }

const INITIAL_STATE: VSCEState = { status: 'search-home' }

export function createVSCEStateMachine(): VSCEStateMachine {
    const states = new BehaviorSubject<VSCEState>(INITIAL_STATE)

    function reducer(state: VSCEState, event: VSCEEvent): VSCEState {
        if (event.type === 'sourcegraph_url_change') {
            return { status: 'context-invalidated' }
        }

        // TODO: hierarchical state, but represented in flat manner.
        // So for example:
        // search
        // - home
        // - results
        // becomes search-home and search-results.
        // Why? To delay bringing in statechart libraries as long as we can.

        switch (state.status) {
            case 'search-home':
                switch (event.type) {
                    case 'submit_search_query':
                        return {
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
        observeState: () => {
            return states.asObservable()
        },
        emit: event => {
            const nextState = reducer(states.value, event)
            states.next(nextState)
        },
    }
}

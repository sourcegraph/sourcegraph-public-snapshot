import { throwError } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import { Client, ClientState } from '../client/client'

/** Reports the client's current state. */
export function getClientState(client: Client): ClientState {
    let clientState: ClientState | undefined
    client.state
        .pipe(first())
        .subscribe(state => (clientState = state))
        .unsubscribe()
    if (clientState === undefined) {
        // This should never happen, because client.state is implemented by a BehaviorSubject that always has a
        // current value.
        throw new Error('client state is not synchronously available')
    }
    return clientState
}

/**
 * Returns a Promise that resolves when the client enters ClientState.Active and rejects if it enters a client
 * state that indicates an error.
 */
export function clientStateIsActive(client: Client): Promise<void> {
    return clientStateIs(client, ClientState.Active, [
        ClientState.ActivateFailed,
        ClientState.ShuttingDown,
        ClientState.Stopped,
    ]).then(() => void 0)
}

/**
 * Returns a Promise that resolves when the client enters the given client state and rejects if it enters one of
 * the reject states.
 */
export function clientStateIs(
    client: Client,
    resolveState: ClientState,
    rejectStates: ClientState[] = []
): Promise<ClientState> {
    return client.state
        .pipe(
            switchMap(state => {
                if (state === resolveState) {
                    return [state]
                }
                if (rejectStates.includes(state)) {
                    return throwError(new Error(`client entered reject state ${ClientState[state]}`))
                }
                return []
            }),
            first()
        )
        .toPromise()
}

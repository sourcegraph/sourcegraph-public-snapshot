import { throwError } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import { Duplex } from 'stream'
import { Client, ClientState } from '../../client/client'
import { MessageTransports } from '../../jsonrpc2/connection'
import { StreamMessageReader, StreamMessageWriter } from '../../jsonrpc2/transports/stream'

/** A bidirectional pipe. */
class TestStream extends Duplex {
    public _write(chunk: string, _encoding: string, done: () => void): void {
        this.emit('data', chunk)
        done()
    }
    public _read(_size: number): void {
        /* noop */
    }
}

/**
 * Creates a pair of message transports that are connected to each other. One can be used as the server and the
 * other as the client.
 */
export function createMessageTransports(): [MessageTransports, MessageTransports] {
    const up = new TestStream()
    const down = new TestStream()
    return [
        { reader: new StreamMessageReader(up), writer: new StreamMessageWriter(down) },
        { reader: new StreamMessageReader(down), writer: new StreamMessageWriter(up) },
    ]
}

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
    return client.state
        .pipe(
            switchMap(state => {
                switch (state) {
                    case ClientState.Initial:
                    case ClientState.Starting:
                    case ClientState.Initializing:
                        return []

                    case ClientState.Active:
                        return [void 0]

                    case ClientState.StartFailed:
                    case ClientState.Stopping:
                    case ClientState.Stopped:
                        return throwError(new Error(`client entered unexpected state ${ClientState[state]}`))
                }
            }),
            first()
        )
        .toPromise()
}

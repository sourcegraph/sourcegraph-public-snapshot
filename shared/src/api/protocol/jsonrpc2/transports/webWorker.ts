import { MessageTransports } from '../connection'
import { Message } from '../messages'
import { AbstractMessageReader, AbstractMessageWriter, DataCallback, MessageReader, MessageWriter } from '../transport'

// TODO: use transferable objects in postMessage for perf

// Copied subset of Worker from the TypeScript "dom"/"webworker" core libraries to avoid needing to add those libs
// to tsconfig.json.
interface MessageEvent {
    data: any
}
interface WorkerEventMap {
    error: any
    message: MessageEvent
}
interface Worker {
    postMessage(message: any): void
    addEventListener<K extends keyof WorkerEventMap>(
        type: K,
        listener: (this: Worker, ev: WorkerEventMap[K]) => any
    ): void
    close?(): void
    terminate?(): void
}

class WebWorkerMessageReader extends AbstractMessageReader implements MessageReader {
    private pending: Message[] = []
    private callback: DataCallback | null = null

    constructor(private worker: Worker) {
        super()

        worker.addEventListener('message', (e: MessageEvent) => {
            try {
                this.processMessage(e)
            } catch (err) {
                this.fireError(err)
            }
        })
        worker.addEventListener('error', err => {
            this.fireError(err)
            terminateWorker(worker)
            this.fireClose()
        })
    }

    private processMessage(e: MessageEvent): void {
        const message: Message = e.data
        if (this.callback) {
            this.callback(message)
        } else {
            this.pending.push(message)
        }
    }

    public listen(callback: DataCallback): void {
        if (this.callback) {
            throw new Error('callback is already set')
        }
        this.callback = callback
        while (this.pending.length !== 0) {
            callback(this.pending.pop()!)
        }
    }

    public unsubscribe(): void {
        super.unsubscribe()
        this.callback = null
        terminateWorker(this.worker)
    }
}

class WebWorkerMessageWriter extends AbstractMessageWriter implements MessageWriter {
    private errorCount = 0

    constructor(private worker: Worker) {
        super()
    }

    public write(message: Message): void {
        try {
            this.worker.postMessage(message)
        } catch (error) {
            this.fireError(error, message, ++this.errorCount)
        }
    }

    public unsubscribe(): void {
        super.unsubscribe()
        terminateWorker(this.worker)
    }
}

function terminateWorker(worker: Worker): void {
    if (worker.terminate) {
        worker.terminate() // in window (worker parent) scope
    } else if (worker.close) {
        worker.close() // in worker scope
    }
}

/**
 * Creates JSON-RPC2 message transports for the Web Worker message communication interface.
 *
 * @param worker The Worker to communicate with (e.g., created with `new Worker(...)`), or the global scope (i.e.,
 *               `self`) if the current execution context is in a Worker. Defaults to the global scope.
 */
export function createWebWorkerMessageTransports(worker: Worker = globalWorkerScope()): MessageTransports {
    return {
        reader: new WebWorkerMessageReader(worker),
        writer: new WebWorkerMessageWriter(worker),
    }
}

function globalWorkerScope(): Worker {
    const worker: Worker = global as any
    // tslint:disable-next-line no-unbound-method
    if (!worker.postMessage || 'document' in worker) {
        throw new Error('global scope is not a Worker')
    }
    return worker
}

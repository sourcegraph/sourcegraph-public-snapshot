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
export interface Worker {
    postMessage(message: any): void
    addEventListener<K extends keyof WorkerEventMap>(
        type: K,
        listener: (this: Worker, ev: WorkerEventMap[K]) => any
    ): void
}

class WebWorkerMessageReader extends AbstractMessageReader implements MessageReader {
    private pending: Message[] = []
    private callback: DataCallback | null = null

    constructor(worker: Worker) {
        super()

        worker.addEventListener('message', (e: MessageEvent) => {
            try {
                this.processMessage(e)
            } catch (err) {
                this.fireError(err)
            }
        })
        worker.addEventListener('error', err => this.fireError(err))
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

    public stop(): void {
        this.callback = null
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
}

/** Creates JSON-RPC2 message transports for the Web Worker message communication interface. */
export function createWebWorkerMessageTransports(worker: Worker): MessageTransports {
    return {
        reader: new WebWorkerMessageReader(worker),
        writer: new WebWorkerMessageWriter(worker),
    }
}

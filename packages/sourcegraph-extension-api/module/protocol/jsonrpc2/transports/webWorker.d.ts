import { MessageTransports } from '../connection';
interface MessageEvent {
    data: any;
}
interface WorkerEventMap {
    error: any;
    message: MessageEvent;
}
interface Worker {
    postMessage(message: any): void;
    addEventListener<K extends keyof WorkerEventMap>(type: K, listener: (this: Worker, ev: WorkerEventMap[K]) => any): void;
    close?(): void;
    terminate?(): void;
}
/**
 * Creates JSON-RPC2 message transports for the Web Worker message communication interface.
 *
 * @param worker The Worker to communicate with (e.g., created with `new Worker(...)`), or the global scope (i.e.,
 *               `self`) if the current execution context is in a Worker. Defaults to the global scope.
 */
export declare function createWebWorkerMessageTransports(worker?: Worker): MessageTransports;
export {};

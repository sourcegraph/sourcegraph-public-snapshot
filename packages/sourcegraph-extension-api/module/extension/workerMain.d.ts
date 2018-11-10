interface MessageEvent {
    data: any;
    origin: string | null;
}
/**
 * This is a subset of DedicatedWorkerGlobalScope. We can't use `/// <references lib="webworker"/>` because
 * Prettier does not support triple-slash directive syntax.
 */
interface DedicatedWorkerGlobalScope {
    location: {
        origin: string;
    };
    addEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
    removeEventListener(type: 'message', listener: (event: MessageEvent) => void): void;
    importScripts(url: string): void;
    close(): void;
}
/**
 * The entrypoint for Web Workers that are spawned to run an extension.
 *
 * To initialize the worker, the parent sends it a message whose data is an object conforming to the
 * {@link InitData} interface. Among other things, this contains the URL of the extension's JavaScript bundle.
 *
 * @param self The worker's `self` global scope.
 */
export declare function extensionHostWorkerMain(self: DedicatedWorkerGlobalScope): void;
export {};

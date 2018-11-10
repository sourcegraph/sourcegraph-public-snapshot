import * as sourcegraph from 'sourcegraph';
import { MessageTransports } from '../protocol/jsonrpc2/connection';
/**
 * Required information when initializing an extension host.
 */
export interface InitData {
    /** The URL to the JavaScript source file (that exports an `activate` function) for the extension. */
    bundleURL: string;
    /** @see {@link module:sourcegraph.internal.sourcegraphURL} */
    sourcegraphURL: string;
    /** @see {@link module:sourcegraph.internal.clientApplication} */
    clientApplication: 'sourcegraph' | 'other';
}
/**
 * Creates the Sourcegraph extension host and the extension API handle (which extensions access with `import
 * sourcegraph from 'sourcegraph'`).
 *
 * @param initData The information to initialize this extension host.
 * @param transports The message reader and writer to use for communication with the client. Defaults to
 *                   communicating using self.postMessage and MessageEvents with the parent (assuming that it is
 *                   called in a Web Worker).
 * @return The extension API.
 */
export declare function createExtensionHost(initData: InitData, transports?: MessageTransports): typeof sourcegraph;

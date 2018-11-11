import { Connection } from '../protocol/jsonrpc2/connection';
/**
 * @returns A proxy that translates method calls on itself to requests sent on the {@link connection}.
 */
export declare function createProxyAndHandleRequests(prefix: string, connection: Connection, handler: any): any;
/**
 * Creates a Proxy that translates method calls (whose name begins with "$") on the returned object to invocations
 * of the {@link call} function with the method name and arguments of the original call.
 */
export declare function createProxy(call: (name: string, args: any[]) => any): any;
/**
 * Forwards all requests received on the connection to the corresponding method on the handler object. The
 * connection method `${prefix}/${name}` corresponds to the `${name}` method on the handler object. names.
 *
 * @param handler - An instance of a class whose methods should be called when the connection receives
 *                  corresponding requests.
 */
export declare function handleRequests(connection: Connection, prefix: string, handler: any): void;

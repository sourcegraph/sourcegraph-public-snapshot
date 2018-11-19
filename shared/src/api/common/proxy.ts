import { Connection } from '../protocol/jsonrpc2/connection'

/**
 * @returns A proxy that translates method calls on itself to requests sent on the {@link connection}.
 */
export function createProxyAndHandleRequests(prefix: string, connection: Connection, handler: any): any {
    handleRequests(connection, prefix, handler)
    return createProxy(connection, prefix)
}

/**
 * Creates a Proxy that translates method calls (whose name begins with "$") on the returned object to messages
 * named `${prefix}/${name}` on the connection.
 *
 * @param connection - The connection to send messages on when proxy methods are called.
 * @param prefix - The name prefix for connection methods.
 */
export function createProxy(connection: Connection, prefix: string): any {
    return new Proxy(Object.create(null), {
        get: (target: any, name: string) => {
            if (!target[name] && name[0] === '$') {
                const method = `${prefix}/${name}`
                if (name.startsWith('$observe')) {
                    target[name] = (...args: any[]) => connection.observeRequest(method, args)
                } else {
                    target[name] = (...args: any[]) => connection.sendRequest(method, args)
                }
            }
            return target[name]
        },
    })
}

/**
 * Forwards all requests received on the connection to the corresponding method on the handler object. The
 * connection method `${prefix}/${name}` corresponds to the `${name}` method on the handler object.
 *
 * @param handler - An instance of a class whose methods should be called when the connection receives
 *                  corresponding requests, or an object created with Object.create(null) (or otherwise with a null
 *                  prototype) whose properties contain functions to be called.
 */
export function handleRequests(connection: Connection, prefix: string, handler: any): void {
    // A class instance's methods are own, non-enumerable properties of its prototype.
    const proto = Object.getPrototypeOf(handler) || handler
    for (const name of Object.getOwnPropertyNames(proto)) {
        const value = proto[name]
        if (name[0] === '$' && typeof value === 'function') {
            connection.onRequest(`${prefix}/${name}`, (args: any[]) => value.apply(handler, args))
        }
    }
}

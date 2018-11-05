import { Connection } from '../protocol/jsonrpc2/connection'

/**
 * @returns A proxy that translates method calls on itself to requests sent on the {@link connection}.
 */
export function createProxyAndHandleRequests(prefix: string, connection: Connection, handler: any): any {
    handleRequests(connection, prefix, handler)
    return createProxy((name, ...args: any[]) => connection.sendRequest(`${prefix}/${name}`, ...args))
}

/**
 * Creates a Proxy that translates method calls (whose name begins with "$") on the returned object to invocations
 * of the {@link call} function with the method name and arguments of the original call.
 */
export function createProxy(call: (name: string, args: any[]) => any): any {
    return new Proxy(Object.create(null), {
        get: (target: any, name: string) => {
            if (!target[name] && name[0] === '$') {
                target[name] = (...args: any[]) => call(name, args)
            }
            return target[name]
        },
    })
}

/**
 * Forwards all requests received on the connection to the corresponding method on the handler object. The
 * connection method `${prefix}/${name}` corresponds to the `${name}` method on the handler object. names.
 *
 * @param handler - An instance of a class whose methods should be called when the connection receives
 *                  corresponding requests.
 */
export function handleRequests(connection: Connection, prefix: string, handler: any): void {
    // A class instance's methods are own, non-enumerable properties of its prototype.
    const proto = Object.getPrototypeOf(handler)
    for (const name of Object.getOwnPropertyNames(proto)) {
        const value = proto[name]
        if (name[0] === '$' && typeof value === 'function') {
            connection.onRequest(`${prefix}/${name}`, (...args: any[]) => value.apply(handler, args[0]))
        }
    }
}

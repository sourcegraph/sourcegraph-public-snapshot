/**
 * @returns A proxy that translates method calls on itself to requests sent on the {@link connection}.
 */
export function createProxyAndHandleRequests(prefix, connection, handler) {
    handleRequests(connection, prefix, handler);
    return createProxy((name, ...args) => connection.sendRequest(`${prefix}/${name}`, ...args));
}
/**
 * Creates a Proxy that translates method calls (whose name begins with "$") on the returned object to invocations
 * of the {@link call} function with the method name and arguments of the original call.
 */
export function createProxy(call) {
    return new Proxy(Object.create(null), {
        get: (target, name) => {
            if (!target[name] && name[0] === '$') {
                target[name] = (...args) => call(name, args);
            }
            return target[name];
        },
    });
}
/**
 * Forwards all requests received on the connection to the corresponding method on the handler object. The
 * connection method `${prefix}/${name}` corresponds to the `${name}` method on the handler object. names.
 *
 * @param handler - An instance of a class whose methods should be called when the connection receives
 *                  corresponding requests.
 */
export function handleRequests(connection, prefix, handler) {
    // A class instance's methods are own, non-enumerable properties of its prototype.
    const proto = Object.getPrototypeOf(handler);
    for (const name of Object.getOwnPropertyNames(proto)) {
        const value = proto[name];
        if (name[0] === '$' && typeof value === 'function') {
            connection.onRequest(`${prefix}/${name}`, (...args) => value.apply(handler, args[0]));
        }
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoicHJveHkuanMiLCJzb3VyY2VSb290Ijoic3JjLyIsInNvdXJjZXMiOlsiY29tbW9uL3Byb3h5LnRzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUVBOztHQUVHO0FBQ0gsTUFBTSxVQUFVLDRCQUE0QixDQUFDLE1BQWMsRUFBRSxVQUFzQixFQUFFLE9BQVk7SUFDN0YsY0FBYyxDQUFDLFVBQVUsRUFBRSxNQUFNLEVBQUUsT0FBTyxDQUFDLENBQUE7SUFDM0MsT0FBTyxXQUFXLENBQUMsQ0FBQyxJQUFJLEVBQUUsR0FBRyxJQUFXLEVBQUUsRUFBRSxDQUFDLFVBQVUsQ0FBQyxXQUFXLENBQUMsR0FBRyxNQUFNLElBQUksSUFBSSxFQUFFLEVBQUUsR0FBRyxJQUFJLENBQUMsQ0FBQyxDQUFBO0FBQ3RHLENBQUM7QUFFRDs7O0dBR0c7QUFDSCxNQUFNLFVBQVUsV0FBVyxDQUFDLElBQXdDO0lBQ2hFLE9BQU8sSUFBSSxLQUFLLENBQUMsTUFBTSxDQUFDLE1BQU0sQ0FBQyxJQUFJLENBQUMsRUFBRTtRQUNsQyxHQUFHLEVBQUUsQ0FBQyxNQUFXLEVBQUUsSUFBWSxFQUFFLEVBQUU7WUFDL0IsSUFBSSxDQUFDLE1BQU0sQ0FBQyxJQUFJLENBQUMsSUFBSSxJQUFJLENBQUMsQ0FBQyxDQUFDLEtBQUssR0FBRyxFQUFFO2dCQUNsQyxNQUFNLENBQUMsSUFBSSxDQUFDLEdBQUcsQ0FBQyxHQUFHLElBQVcsRUFBRSxFQUFFLENBQUMsSUFBSSxDQUFDLElBQUksRUFBRSxJQUFJLENBQUMsQ0FBQTthQUN0RDtZQUNELE9BQU8sTUFBTSxDQUFDLElBQUksQ0FBQyxDQUFBO1FBQ3ZCLENBQUM7S0FDSixDQUFDLENBQUE7QUFDTixDQUFDO0FBRUQ7Ozs7OztHQU1HO0FBQ0gsTUFBTSxVQUFVLGNBQWMsQ0FBQyxVQUFzQixFQUFFLE1BQWMsRUFBRSxPQUFZO0lBQy9FLGtGQUFrRjtJQUNsRixNQUFNLEtBQUssR0FBRyxNQUFNLENBQUMsY0FBYyxDQUFDLE9BQU8sQ0FBQyxDQUFBO0lBQzVDLEtBQUssTUFBTSxJQUFJLElBQUksTUFBTSxDQUFDLG1CQUFtQixDQUFDLEtBQUssQ0FBQyxFQUFFO1FBQ2xELE1BQU0sS0FBSyxHQUFHLEtBQUssQ0FBQyxJQUFJLENBQUMsQ0FBQTtRQUN6QixJQUFJLElBQUksQ0FBQyxDQUFDLENBQUMsS0FBSyxHQUFHLElBQUksT0FBTyxLQUFLLEtBQUssVUFBVSxFQUFFO1lBQ2hELFVBQVUsQ0FBQyxTQUFTLENBQUMsR0FBRyxNQUFNLElBQUksSUFBSSxFQUFFLEVBQUUsQ0FBQyxHQUFHLElBQVcsRUFBRSxFQUFFLENBQUMsS0FBSyxDQUFDLEtBQUssQ0FBQyxPQUFPLEVBQUUsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQTtTQUMvRjtLQUNKO0FBQ0wsQ0FBQyIsInNvdXJjZXNDb250ZW50IjpbImltcG9ydCB7IENvbm5lY3Rpb24gfSBmcm9tICcuLi9wcm90b2NvbC9qc29ucnBjMi9jb25uZWN0aW9uJ1xuXG4vKipcbiAqIEByZXR1cm5zIEEgcHJveHkgdGhhdCB0cmFuc2xhdGVzIG1ldGhvZCBjYWxscyBvbiBpdHNlbGYgdG8gcmVxdWVzdHMgc2VudCBvbiB0aGUge0BsaW5rIGNvbm5lY3Rpb259LlxuICovXG5leHBvcnQgZnVuY3Rpb24gY3JlYXRlUHJveHlBbmRIYW5kbGVSZXF1ZXN0cyhwcmVmaXg6IHN0cmluZywgY29ubmVjdGlvbjogQ29ubmVjdGlvbiwgaGFuZGxlcjogYW55KTogYW55IHtcbiAgICBoYW5kbGVSZXF1ZXN0cyhjb25uZWN0aW9uLCBwcmVmaXgsIGhhbmRsZXIpXG4gICAgcmV0dXJuIGNyZWF0ZVByb3h5KChuYW1lLCAuLi5hcmdzOiBhbnlbXSkgPT4gY29ubmVjdGlvbi5zZW5kUmVxdWVzdChgJHtwcmVmaXh9LyR7bmFtZX1gLCAuLi5hcmdzKSlcbn1cblxuLyoqXG4gKiBDcmVhdGVzIGEgUHJveHkgdGhhdCB0cmFuc2xhdGVzIG1ldGhvZCBjYWxscyAod2hvc2UgbmFtZSBiZWdpbnMgd2l0aCBcIiRcIikgb24gdGhlIHJldHVybmVkIG9iamVjdCB0byBpbnZvY2F0aW9uc1xuICogb2YgdGhlIHtAbGluayBjYWxsfSBmdW5jdGlvbiB3aXRoIHRoZSBtZXRob2QgbmFtZSBhbmQgYXJndW1lbnRzIG9mIHRoZSBvcmlnaW5hbCBjYWxsLlxuICovXG5leHBvcnQgZnVuY3Rpb24gY3JlYXRlUHJveHkoY2FsbDogKG5hbWU6IHN0cmluZywgYXJnczogYW55W10pID0+IGFueSk6IGFueSB7XG4gICAgcmV0dXJuIG5ldyBQcm94eShPYmplY3QuY3JlYXRlKG51bGwpLCB7XG4gICAgICAgIGdldDogKHRhcmdldDogYW55LCBuYW1lOiBzdHJpbmcpID0+IHtcbiAgICAgICAgICAgIGlmICghdGFyZ2V0W25hbWVdICYmIG5hbWVbMF0gPT09ICckJykge1xuICAgICAgICAgICAgICAgIHRhcmdldFtuYW1lXSA9ICguLi5hcmdzOiBhbnlbXSkgPT4gY2FsbChuYW1lLCBhcmdzKVxuICAgICAgICAgICAgfVxuICAgICAgICAgICAgcmV0dXJuIHRhcmdldFtuYW1lXVxuICAgICAgICB9LFxuICAgIH0pXG59XG5cbi8qKlxuICogRm9yd2FyZHMgYWxsIHJlcXVlc3RzIHJlY2VpdmVkIG9uIHRoZSBjb25uZWN0aW9uIHRvIHRoZSBjb3JyZXNwb25kaW5nIG1ldGhvZCBvbiB0aGUgaGFuZGxlciBvYmplY3QuIFRoZVxuICogY29ubmVjdGlvbiBtZXRob2QgYCR7cHJlZml4fS8ke25hbWV9YCBjb3JyZXNwb25kcyB0byB0aGUgYCR7bmFtZX1gIG1ldGhvZCBvbiB0aGUgaGFuZGxlciBvYmplY3QuIG5hbWVzLlxuICpcbiAqIEBwYXJhbSBoYW5kbGVyIC0gQW4gaW5zdGFuY2Ugb2YgYSBjbGFzcyB3aG9zZSBtZXRob2RzIHNob3VsZCBiZSBjYWxsZWQgd2hlbiB0aGUgY29ubmVjdGlvbiByZWNlaXZlc1xuICogICAgICAgICAgICAgICAgICBjb3JyZXNwb25kaW5nIHJlcXVlc3RzLlxuICovXG5leHBvcnQgZnVuY3Rpb24gaGFuZGxlUmVxdWVzdHMoY29ubmVjdGlvbjogQ29ubmVjdGlvbiwgcHJlZml4OiBzdHJpbmcsIGhhbmRsZXI6IGFueSk6IHZvaWQge1xuICAgIC8vIEEgY2xhc3MgaW5zdGFuY2UncyBtZXRob2RzIGFyZSBvd24sIG5vbi1lbnVtZXJhYmxlIHByb3BlcnRpZXMgb2YgaXRzIHByb3RvdHlwZS5cbiAgICBjb25zdCBwcm90byA9IE9iamVjdC5nZXRQcm90b3R5cGVPZihoYW5kbGVyKVxuICAgIGZvciAoY29uc3QgbmFtZSBvZiBPYmplY3QuZ2V0T3duUHJvcGVydHlOYW1lcyhwcm90bykpIHtcbiAgICAgICAgY29uc3QgdmFsdWUgPSBwcm90b1tuYW1lXVxuICAgICAgICBpZiAobmFtZVswXSA9PT0gJyQnICYmIHR5cGVvZiB2YWx1ZSA9PT0gJ2Z1bmN0aW9uJykge1xuICAgICAgICAgICAgY29ubmVjdGlvbi5vblJlcXVlc3QoYCR7cHJlZml4fS8ke25hbWV9YCwgKC4uLmFyZ3M6IGFueVtdKSA9PiB2YWx1ZS5hcHBseShoYW5kbGVyLCBhcmdzWzBdKSlcbiAgICAgICAgfVxuICAgIH1cbn1cbiJdfQ==
// polyfill MessagePort and MessageChannel
export class AsyncMessagePort implements MessagePort {
    public onmessage = null
    public onmessageerror = null

    public otherPort: AsyncMessagePort = null
    private onmessageListeners: EventListener[] = []

    constructor() {}

    public dispatchEvent(event) {
        if (this.onmessage) {
            this.onmessage(event)
        }
        this.onmessageListeners.forEach(listener => listener(event))
        return true
    }

    public postMessage(message) {
        if (!this.otherPort) {
            return
        }
        this.otherPort.dispatchEvent({ data: message })
        // setTimeout(() => this.otherPort.dispatchEvent({ data: message }))
    }

    public addEventListener(type, listener) {
        if (type !== 'message') {
            return
        }
        if (typeof listener !== 'function' || this.onmessageListeners.indexOf(listener) !== -1) {
            return
        }
        this.onmessageListeners.push(listener)
    }

    public removeEventListener(type, listener) {
        if (type !== 'message') {
            return
        }
        const index = this.onmessageListeners.indexOf(listener)
        if (index === -1) {
            return
        }

        this.onmessageListeners.splice(index, 1)
    }

    public start() {
        // do nothing at this moment
    }

    public close() {
        // do nothing at this moment
    }
}

export class AsyncMessageChannel implements MessageChannel {
    public port1: AsyncMessagePort
    public port2: AsyncMessagePort
    constructor() {
        this.port1 = new AsyncMessagePort()
        this.port2 = new AsyncMessagePort()
        this.port1.otherPort = this.port2
        this.port2.otherPort = this.port1
    }
}

/**
 * https://github.com/zloirock/core-js/blob/master/packages/core-js/internals/global.js
 */
const globalObj =
    typeof window !== 'undefined' && (window as any).Math === Math
        ? window
        : typeof self !== 'undefined' && (self as any).Math === Math
        ? self
        : Function('return this')()

export function applyPolyfill() {
    globalObj.MessagePort = AsyncMessagePort
    globalObj.MessageChannel = AsyncMessageChannel
}

if (!globalObj.MessagePort || !globalObj.MessageChannel) {
    applyPolyfill()
}

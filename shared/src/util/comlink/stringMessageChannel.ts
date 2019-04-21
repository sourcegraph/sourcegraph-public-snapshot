export interface StringMessagePort {
    send(data: string): void
    addListener(listener: (data: string) => void): void
    removeListener(listener: (data: string) => void): void
}

interface Message {
    id: string
    msg: any
    messageChannels: string[][]
}

// smc is the Chrome port that lets the client communicate with the ext host.
//
// Port1 sends to the exthost and receives from the exthost (it's like the original smc).
//
// Port2 sends to the client and receives from the client, so port2.onmessage = from client,
// port2.postMessage = to client.
export function wrapStringMessagePort(smc: StringMessagePort, id: string | null = null): MessagePort {
    const { port1, port2 } = new MessageChannel()
    hookupSMC(port2, smc, id)
    return port1
}

function hookupSMC(internalPort: MessagePort, smc: StringMessagePort, id: string | null = null): void {
    // Intercept messages sent from the client to the exthost before they are sent to the exthost.
    // For all MessageChannels recursively in the message, replace them with a sentinel indicating
    // that their value is a new virtual channel.
    internalPort.onmessage = (event: MessageEvent) => {
        const msg = event.data
        const messageChannels = Array.from(findMessageChannels(event.data))
        for (const messageChannel of messageChannels) {
            const id = generateUID()
            const channel: MessagePort = replaceProperty(msg, messageChannel, id)
            const origClose = channel.close.bind(channel)
            channel.close = () => {
                origClose()
                console.log('CLOSED!!!!!')
            }
            hookupSMC(channel, smc, id)
        }
        const payload = JSON.stringify({ id, msg, messageChannels })
        smc.send(payload)
    }

    // Intercept messages sent from the exthost to the client before they are received by the client. For all messageChannels mentioned in the data,
    smc.addListener(
        (dataStr): void => {
            const data: Message = JSON.parse(dataStr)
            if (id && id !== data.id) {
                return
            }
            if (!id && data.id) {
                return
            }
            const mcs = data.messageChannels.map(messageChannel => {
                const id = messageChannel.reduce((obj, key) => obj[key], data.msg)
                const port = wrapStringMessagePort(smc, id) // create a sub-channel
                replaceProperty(data.msg, messageChannel, port)
                return port
            })
            internalPort.postMessage(data.msg, mcs)
        }
    )
}

function replaceProperty(obj: any, path: string[], newVal: any): any {
    for (const key of path.slice(0, -1)) {
        obj = obj[key]
    }
    const key = path[path.length - 1]
    const orig = obj[key]
    obj[key] = newVal
    return orig
}

function* findMessageChannels(obj: any, path: string[] = []): Iterable<string[]> {
    if (!obj) {
        return
    }
    if (typeof obj === 'string') {
        return
    }
    if (obj instanceof MessagePort) {
        yield path.slice()
        return
    }
    for (const key of Object.keys(obj)) {
        path.push(key)
        yield* findMessageChannels(obj[key], path)
        path.pop()
    }
}

function hex4(): string {
    return Math.floor((1 + Math.random()) * 0x10000)
        .toString(16)
        .substring(1)
}

const bits = 128
function generateUID(): string {
    return new Array(bits / 16)
        .fill(0)
        .map(_ => hex4())
        .join('')
}

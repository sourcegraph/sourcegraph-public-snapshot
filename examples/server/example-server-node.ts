import * as net from 'net'
import * as stream from 'stream'
import * as WebSocket from 'ws'
import { TextDocumentPublishDecorationsNotification } from '../../lib/protocol'
import { createWebSocketMessageTransports } from '../../src/jsonrpc2/transports/nodeWebSocket'
import { StreamMessageReader, StreamMessageWriter } from '../../src/jsonrpc2/transports/stream'
import { InitializeResult } from '../../src/protocol'
import { TextDocumentDecoration } from '../../src/protocol/decoration'
import { Connection, createConnection } from '../../src/server/server'

function register(connection: Connection): void {
    connection.onInitialize(
        () =>
            ({
                capabilities: {
                    textDocumentSync: { openClose: true },
                    decorationProvider: true,
                },
            } as InitializeResult)
    )

    connection.onDidOpenTextDocument(params =>
        connection.sendNotification(TextDocumentPublishDecorationsNotification.type, {
            textDocument: params.textDocument,
            decorations: ['cyan', 'magenta', 'yellow', 'black', 'cyan', 'magenta', 'yellow', 'black'].map(
                (color, i) =>
                    ({
                        range: { start: { line: i, character: 0 }, end: { line: i, character: 0 } },
                        isWholeLine: true,
                        backgroundColor: color,
                    } as TextDocumentDecoration)
            ),
        })
    )
}

const log = (message?: string, ...args: any[]) => console.error(message, ...args)

function addr(): { host: string; port: number } {
    const addr = process.env.CX_ADDR || ''
    const [host, portStr] = addr.split(':', 2)
    if (!host || !portStr) {
        log('Invalid CX_ADDR.')
        process.exit(1)
    }
    const port = parseInt(portStr, 10)
    if (Number.isNaN(port)) {
        log('Invalid CX_ADDR port.')
        process.exit(1)
    }
    return { host, port }
}

switch (process.env.CX_MODE) {
    case 'stdio': {
        log('Starting CXP connection on stdio.')
        const connection = createConnection({
            reader: new StreamMessageReader(process.stdin),
            writer: new StreamMessageWriter(process.stdout),
        })
        register(connection)
        connection.listen()
        log('Listening on stdin and responding on stdout')
        break
    }

    case 'tcp': {
        const { host, port } = addr()
        net.createServer(socket => {
            const reader = new stream.PassThrough()
            const writer = new stream.PassThrough()

            socket.pipe(reader)
            writer.pipe(socket)

            const connection = createConnection({
                reader: new StreamMessageReader(reader),
                writer: new StreamMessageWriter(writer),
            })
            register(connection)
            connection.listen()
        })
            .listen({ host, port })
            .on('listening', () => log(`Listening for CXP connections on TCP. addr=${host}:${port}`))
            .on('connection', () => log('Client connected.'))
            .on('error', err => log('Error: ', err))
        break
    }

    case 'websocket': {
        const { host, port } = addr()
        new WebSocket.Server({ host, port })
            .on('listening', () => log(`Listening for CXP connections over WebSockets. addr=${host}:${port}`))
            .on('connection', ws => {
                log('Client connected.')
                createWebSocketMessageTransports(ws).then(transports => {
                    const connection = createConnection(transports)
                    register(connection)
                    connection.listen()
                }, log)
            })
            .on('error', err => log('Error: ', err))
        break
    }

    default:
        log('Invalid CX_MODE.')
        process.exit(1)
}

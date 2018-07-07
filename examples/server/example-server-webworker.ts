import { createWebWorkerMessageTransports, Worker } from '../../src/jsonrpc2/transports/webWorker'
import { InitializeResult } from '../../src/protocol'
import { TextDocumentDecoration, TextDocumentDecorationsParams } from '../../src/protocol/decorations'
import { Connection, createConnection } from '../../src/server/server'

declare var self: Worker

function register(connection: Connection): void {
    connection.onInitialize(
        params =>
            ({
                capabilities: { decorationsProvider: { static: true } },
            } as InitializeResult)
    )

    connection.onRequest(
        'textDocument/decorations',
        (params: TextDocumentDecorationsParams): TextDocumentDecoration[] =>
            ['cyan', 'magenta', 'yellow', 'black', 'cyan', 'magenta', 'yellow', 'black'].map(
                (color, i) =>
                    ({
                        range: { start: { line: i, character: 0 }, end: { line: i, character: 0 } },
                        isWholeLine: true,
                        backgroundColor: color,
                    } as TextDocumentDecoration)
            )
    )
}

const connection = createConnection(createWebWorkerMessageTransports(self))
register(connection)
connection.listen()

import { TextDocumentPublishDecorationsNotification } from '../../lib/protocol'
import { createWebWorkerMessageTransports, Worker } from '../../src/jsonrpc2/transports/webWorker'
import { InitializeResult } from '../../src/protocol'
import { TextDocumentDecoration } from '../../src/protocol/decoration'
import { Connection, createConnection } from '../../src/server/server'

declare var self: Worker

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

const connection = createConnection(createWebWorkerMessageTransports(self))
register(connection)
connection.listen()

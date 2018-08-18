import {
    TextDocument,
    TextDocumentChangeEvent,
    TextDocumentContentChangeEvent,
    TextDocumentWillSaveEvent,
    TextEdit,
} from 'vscode-languageserver-types'
import { CancellationToken } from '../../jsonrpc2/cancel'
import { Emitter, Event } from '../../jsonrpc2/events'
import { RequestHandler } from '../../jsonrpc2/handlers'
import {
    DidChangeTextDocumentParams,
    DidCloseTextDocumentParams,
    DidOpenTextDocumentParams,
    DidSaveTextDocumentParams,
    TextDocumentSyncKind,
    WillSaveTextDocumentParams,
} from '../../protocol'
import { isFunction } from '../../util'
import { Connection } from '../server'

/**
 * A manager for simple text documents
 */
export class TextDocuments {
    private _documents: { [uri: string]: TextDocument }

    private _onDidChangeContent: Emitter<TextDocumentChangeEvent>
    private _onDidOpen: Emitter<TextDocumentChangeEvent>
    private _onDidClose: Emitter<TextDocumentChangeEvent>
    private _onDidSave: Emitter<TextDocumentChangeEvent>
    private _onWillSave: Emitter<TextDocumentWillSaveEvent>
    private _willSaveWaitUntil?: RequestHandler<TextDocumentWillSaveEvent, TextEdit[], void>

    /**
     * Create a new text document manager.
     */
    public constructor() {
        this._documents = Object.create(null)
        this._onDidChangeContent = new Emitter<TextDocumentChangeEvent>()
        this._onDidOpen = new Emitter<TextDocumentChangeEvent>()
        this._onDidClose = new Emitter<TextDocumentChangeEvent>()
        this._onDidSave = new Emitter<TextDocumentChangeEvent>()
        this._onWillSave = new Emitter<TextDocumentWillSaveEvent>()
    }

    /**
     * Returns the [TextDocumentSyncKind](#TextDocumentSyncKind) used by
     * this text document manager.
     */
    public get syncKind(): TextDocumentSyncKind {
        return TextDocumentSyncKind.Full
    }

    /**
     * An event that fires when a text document managed by this manager
     * has been opened or the content changes.
     */
    public get onDidChangeContent(): Event<TextDocumentChangeEvent> {
        return this._onDidChangeContent.event
    }

    /**
     * An event that fires when a text document managed by this manager
     * has been opened.
     */
    public get onDidOpen(): Event<TextDocumentChangeEvent> {
        return this._onDidOpen.event
    }

    /**
     * An event that fires when a text document managed by this manager
     * will be saved.
     */
    public get onWillSave(): Event<TextDocumentWillSaveEvent> {
        return this._onWillSave.event
    }

    /**
     * Sets a handler that will be called if a participant wants to provide
     * edits during a text document save.
     */
    public onWillSaveWaitUntil(handler: RequestHandler<TextDocumentWillSaveEvent, TextEdit[], void>): void {
        this._willSaveWaitUntil = handler
    }

    /**
     * An event that fires when a text document managed by this manager
     * has been saved.
     */
    public get onDidSave(): Event<TextDocumentChangeEvent> {
        return this._onDidSave.event
    }

    /**
     * An event that fires when a text document managed by this manager
     * has been closed.
     */
    public get onDidClose(): Event<TextDocumentChangeEvent> {
        return this._onDidClose.event
    }

    /**
     * Returns the document for the given URI. Returns undefined if
     * the document is not mananged by this instance.
     *
     * @param uri The text document's URI to retrieve.
     * @return the text document or `undefined`.
     */
    public get(uri: string): TextDocument | undefined {
        return this._documents[uri]
    }

    /**
     * Returns all text documents managed by this instance.
     *
     * @return all text documents.
     */
    public all(): TextDocument[] {
        return Object.keys(this._documents).map(key => this._documents[key])
    }

    /**
     * Returns the URIs of all text documents managed by this instance.
     *
     * @return the URI's of all text documents.
     */
    public keys(): string[] {
        return Object.keys(this._documents)
    }

    /**
     * Listens for `low level` notification on the given connection to
     * update the text documents managed by this instance.
     *
     * @param connection The connection to listen on.
     */
    public listen(connection: Connection): void {
        interface UpdateableDocument extends TextDocument {
            update(event: TextDocumentContentChangeEvent, version: number): void
        }

        function isUpdateableDocument(value: TextDocument): value is UpdateableDocument {
            // tslint:disable-next-line:no-unbound-method
            return isFunction((value as UpdateableDocument).update)
        }

        connection.onDidOpenTextDocument((event: DidOpenTextDocumentParams) => {
            const td = event.textDocument
            const document = TextDocument.create(td.uri, td.languageId, td.version, td.text)
            this._documents[td.uri] = document
            const toFire = Object.freeze({ document })
            this._onDidOpen.fire(toFire)
            this._onDidChangeContent.fire(toFire)
        })
        connection.onDidChangeTextDocument((event: DidChangeTextDocumentParams) => {
            const td = event.textDocument
            const changes = event.contentChanges
            const last: TextDocumentContentChangeEvent | undefined =
                changes.length > 0 ? changes[changes.length - 1] : undefined
            if (last) {
                const document = this._documents[td.uri]
                if (document && isUpdateableDocument(document)) {
                    if (td.version === null || td.version === void 0) {
                        throw new Error(`Recevied document change event for ${td.uri} without valid version identifier`)
                    }
                    document.update(last, td.version)
                    this._onDidChangeContent.fire(Object.freeze({ document }))
                }
            }
        })
        connection.onDidCloseTextDocument((event: DidCloseTextDocumentParams) => {
            const document = this._documents[event.textDocument.uri]
            if (document) {
                delete this._documents[event.textDocument.uri]
                this._onDidClose.fire(Object.freeze({ document }))
            }
        })
        connection.onWillSaveTextDocument((event: WillSaveTextDocumentParams) => {
            const document = this._documents[event.textDocument.uri]
            if (document) {
                this._onWillSave.fire(Object.freeze({ document, reason: event.reason }))
            }
        })
        connection.onWillSaveTextDocumentWaitUntil((event: WillSaveTextDocumentParams, token: CancellationToken) => {
            const document = this._documents[event.textDocument.uri]
            if (document && this._willSaveWaitUntil) {
                return this._willSaveWaitUntil(Object.freeze({ document, reason: event.reason }), token)
            } else {
                return []
            }
        })
        connection.onDidSaveTextDocument((event: DidSaveTextDocumentParams) => {
            const document = this._documents[event.textDocument.uri]
            if (document) {
                this._onDidSave.fire(Object.freeze({ document }))
            }
        })
    }
}

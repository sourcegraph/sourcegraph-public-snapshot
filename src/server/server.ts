import { Unsubscribable } from 'rxjs'
import {
    CodeAction,
    CodeLens,
    Command,
    CompletionItem,
    CompletionList,
    Definition,
    DocumentHighlight,
    DocumentLink,
    DocumentSymbolParams,
    Hover,
    Location,
    SignatureHelp,
    SymbolInformation,
    TextEdit,
    WorkspaceEdit,
    WorkspaceSymbolParams,
} from 'vscode-languageserver-types'
import { CancellationToken, CancellationTokenSource } from '../jsonrpc2/cancel'
import { createMessageConnection, MessageTransports } from '../jsonrpc2/connection'
import { ConnectionStrategy } from '../jsonrpc2/connectionStrategy'
import {
    GenericNotificationHandler,
    GenericRequestHandler,
    NotificationHandler,
    RequestHandler,
    StarNotificationHandler,
    StarRequestHandler,
} from '../jsonrpc2/handlers'
import { MessageType as RPCMessageType, NotificationType, RequestType, ResponseError } from '../jsonrpc2/messages'
import { SetTraceNotification, Trace } from '../jsonrpc2/trace'
import {
    CodeActionParams,
    CodeActionRequest,
    CodeLensParams,
    CodeLensRequest,
    CodeLensResolveRequest,
    ColorInformation,
    ColorPresentation,
    ColorPresentationParams,
    ColorPresentationRequest,
    CompletionParams,
    CompletionRequest,
    CompletionResolveRequest,
    DefinitionRequest,
    DidChangeConfigurationNotification,
    DidChangeConfigurationParams,
    DidChangeTextDocumentNotification,
    DidChangeTextDocumentParams,
    DidChangeWatchedFilesNotification,
    DidChangeWatchedFilesParams,
    DidCloseTextDocumentNotification,
    DidCloseTextDocumentParams,
    DidOpenTextDocumentNotification,
    DidOpenTextDocumentParams,
    DidSaveTextDocumentNotification,
    DidSaveTextDocumentParams,
    DocumentColorParams,
    DocumentColorRequest,
    DocumentFormattingParams,
    DocumentFormattingRequest,
    DocumentHighlightRequest,
    DocumentLinkParams,
    DocumentLinkRequest,
    DocumentLinkResolveRequest,
    DocumentOnTypeFormattingParams,
    DocumentOnTypeFormattingRequest,
    DocumentRangeFormattingParams,
    DocumentRangeFormattingRequest,
    DocumentSymbolRequest,
    ExecuteCommandParams,
    ExecuteCommandRequest,
    ExitNotification,
    HoverRequest,
    ImplementationRequest,
    InitializedNotification,
    InitializedParams,
    InitializeError,
    InitializeParams,
    InitializeRequest,
    InitializeResult,
    PublishDiagnosticsNotification,
    PublishDiagnosticsParams,
    ReferenceParams,
    ReferencesRequest,
    RenameParams,
    RenameRequest,
    ShutdownRequest,
    SignatureHelpRequest,
    TextDocumentPositionParams,
    TypeDefinitionRequest,
    WillSaveTextDocumentNotification,
    WillSaveTextDocumentParams,
    WillSaveTextDocumentWaitUntilRequest,
    WorkspaceSymbolRequest,
} from '../protocol'
import { RemoteClient, RemoteClientImpl } from './features/client'
import { Remote } from './features/common'
import { ConnectionLogger, RemoteConsole } from './features/console'
import { RemoteContext, RemoteContextImpl } from './features/context'
import { Telemetry, TelemetryImpl } from './features/telemetry'
import { Tracer, TracerImpl } from './features/tracer'
import { RemoteWindow, RemoteWindowImpl } from './features/window'
import { RemoteWorkspace, RemoteWorkspaceImpl } from './features/workspace'

// Copied from vscode-jsonrpc to avoid adding extraneous dependencies.

/**
 * An empty interface for new proposed API.
 */
// tslint:disable-next-line:class-name
export interface _ {}

/**
 * Interface to describe the shape of the server connection.
 */
export interface Connection<
    PConsole = _,
    PContext = _,
    PTracer = _,
    PTelemetry = _,
    PClient = _,
    PWindow = _,
    PWorkspace = _
> extends Unsubscribable {
    /**
     * Start listening on the input stream for messages to process.
     */
    listen(): void

    /**
     * Installs a request handler described by the given [RequestType](#RequestType).
     *
     * @param type The [RequestType](#RequestType) describing the request.
     * @param handler The handler to install
     */
    onRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, handler: RequestHandler<P, R, E>): void

    /**
     * Installs a request handler for the given method.
     *
     * @param method The method to register a request handler for.
     * @param handler The handler to install.
     */
    onRequest<R, E>(method: string, handler: GenericRequestHandler<R, E>): void

    /**
     * Installs a request handler that is invoked if no specific request handler can be found.
     *
     * @param handler a handler that handles all requests.
     */
    onRequest(handler: StarRequestHandler): void

    /**
     * Send a request to the client.
     *
     * @param type The [RequestType](#RequestType) describing the request.
     * @param params The request's parameters.
     */
    sendRequest<P, R, E, RO>(type: RequestType<P, R, E, RO>, params: P, token?: CancellationToken): Promise<R>

    /**
     * Send a request to the client.
     *
     * @param method The method to invoke on the client.
     * @param params The request's parameters.
     */
    sendRequest<R>(method: string, token?: CancellationToken): Promise<R>
    sendRequest<R>(method: string, params: any, token?: CancellationToken): Promise<R>

    /**
     * Installs a notification handler described by the given [NotificationType](#NotificationType).
     *
     * @param type The [NotificationType](#NotificationType) describing the notification.
     * @param handler The handler to install.
     */
    onNotification<P, RO>(type: NotificationType<P, RO>, handler: NotificationHandler<P>): void

    /**
     * Installs a notification handler for the given method.
     *
     * @param method The method to register a request handler for.
     * @param handler The handler to install.
     */
    onNotification(method: string, handler: GenericNotificationHandler): void

    /**
     * Installs a notification handler that is invoked if no specific notification handler can be found.
     *
     * @param handler a handler that handles all notifications.
     */
    onNotification(handler: StarNotificationHandler): void

    /**
     * Send a notification to the client.
     *
     * @param type The [NotificationType](#NotificationType) describing the notification.
     * @param params The notification's parameters.
     */
    sendNotification<P, RO>(type: NotificationType<P, RO>, params: P): void

    /**
     * Send a notification to the client.
     *
     * @param method The method to invoke on the client.
     * @param params The notification's parameters.
     */
    sendNotification(method: string, params?: any): void

    /**
     * Installs a handler for the initialize request.
     *
     * @param handler The initialize handler.
     */
    onInitialize(handler: RequestHandler<InitializeParams, InitializeResult, InitializeError>): void

    /**
     * Installs a handler for the initialized notification.
     *
     * @param handler The initialized handler.
     */
    onInitialized(handler: NotificationHandler<InitializedParams>): void

    /**
     * Installs a handler for the shutdown request.
     *
     * @param handler The initialize handler.
     */
    onShutdown(handler: RequestHandler<null, void, void>): void

    /**
     * Installs a handler for the exit notification.
     *
     * @param handler The exit handler.
     */
    onExit(handler: NotificationHandler<null>): void

    /**
     * A proxy for the development console. See [RemoteConsole](#RemoteConsole)
     */
    console: RemoteConsole & PConsole

    /** A proxy for the client's context. */
    context: RemoteContext & PContext

    /**
     * A proxy to send trace events to the client.
     */
    tracer: Tracer & PTracer

    /**
     * A proxy to send telemetry events to the client.
     */
    telemetry: Telemetry & PTelemetry

    /**
     * A proxy interface for the language client interface to register for requests or
     * notifications.
     */
    client: RemoteClient & PClient

    /**
     * A proxy for the window. See [RemoteWindow](#RemoteWindow)
     */
    window: RemoteWindow & PWindow

    /**
     * A proxy to talk to the client's workspace.
     */
    workspace: RemoteWorkspace & PWorkspace

    /**
     * Installs a handler for the `DidChangeConfiguration` notification.
     *
     * @param handler The corresponding handler.
     */
    onDidChangeConfiguration(handler: NotificationHandler<DidChangeConfigurationParams>): void

    /**
     * Installs a handler for the `DidChangeWatchedFiles` notification.
     *
     * @param handler The corresponding handler.
     */
    onDidChangeWatchedFiles(handler: NotificationHandler<DidChangeWatchedFilesParams>): void

    /**
     * Installs a handler for the `DidOpenTextDocument` notification.
     *
     * @param handler The corresponding handler.
     */
    onDidOpenTextDocument(handler: NotificationHandler<DidOpenTextDocumentParams>): void

    /**
     * Installs a handler for the `DidChangeTextDocument` notification.
     *
     * @param handler The corresponding handler.
     */
    onDidChangeTextDocument(handler: NotificationHandler<DidChangeTextDocumentParams>): void

    /**
     * Installs a handler for the `DidCloseTextDocument` notification.
     *
     * @param handler The corresponding handler.
     */
    onDidCloseTextDocument(handler: NotificationHandler<DidCloseTextDocumentParams>): void

    /**
     * Installs a handler for the `WillSaveTextDocument` notification.
     *
     * Note that this notification is opt-in. The client will not send it unless
     * your server has the `textDocumentSync.willSave` capability or you've
     * dynamically registered for the `textDocument/willSave` method.
     *
     * @param handler The corresponding handler.
     */
    onWillSaveTextDocument(handler: NotificationHandler<WillSaveTextDocumentParams>): void

    /**
     * Installs a handler for the `WillSaveTextDocumentWaitUntil` request.
     *
     * Note that this request is opt-in. The client will not send it unless
     * your server has the `textDocumentSync.willSaveWaitUntil` capability,
     * or you've dynamically registered for the `textDocument/willSaveWaitUntil`
     * method.
     *
     * @param handler The corresponding handler.
     */
    onWillSaveTextDocumentWaitUntil(
        handler: RequestHandler<WillSaveTextDocumentParams, TextEdit[] | undefined | null, void>
    ): void

    /**
     * Installs a handler for the `DidSaveTextDocument` notification.
     *
     * @param handler The corresponding handler.
     */
    onDidSaveTextDocument(handler: NotificationHandler<DidSaveTextDocumentParams>): void

    /**
     * Sends diagnostics computed for a given document to render them in the user interface.
     *
     * @param params The diagnostic parameters.
     */
    sendDiagnostics(params: PublishDiagnosticsParams): void

    /**
     * Installs a handler for the `Hover` request.
     *
     * @param handler The corresponding handler.
     */
    onHover(handler: RequestHandler<TextDocumentPositionParams, Hover | undefined | null, void>): void

    /**
     * Installs a handler for the `Completion` request.
     *
     * @param handler The corresponding handler.
     */
    onCompletion(
        handler: RequestHandler<CompletionParams, CompletionItem[] | CompletionList | undefined | null, void>
    ): void

    /**
     * Installs a handler for the `CompletionResolve` request.
     *
     * @param handler The corresponding handler.
     */
    onCompletionResolve(handler: RequestHandler<CompletionItem, CompletionItem, void>): void

    /**
     * Installs a handler for the `SignatureHelp` request.
     *
     * @param handler The corresponding handler.
     */
    onSignatureHelp(handler: RequestHandler<TextDocumentPositionParams, SignatureHelp | undefined | null, void>): void

    /**
     * Installs a handler for the `Definition` request.
     *
     * @param handler The corresponding handler.
     */
    onDefinition(handler: RequestHandler<TextDocumentPositionParams, Definition | undefined | null, void>): void

    /**
     * Installs a handler for the `Type Definition` request.
     *
     * @param handler The corresponding handler.
     */
    onTypeDefinition(handler: RequestHandler<TextDocumentPositionParams, Definition | undefined | null, void>): void

    /**
     * Installs a handler for the `Implementation` request.
     *
     * @param handler The corresponding handler.
     */
    onImplementation(handler: RequestHandler<TextDocumentPositionParams, Definition | undefined | null, void>): void

    /**
     * Installs a handler for the `References` request.
     *
     * @param handler The corresponding handler.
     */
    onReferences(handler: RequestHandler<ReferenceParams, Location[] | undefined | null, void>): void

    /**
     * Installs a handler for the `DocumentHighlight` request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentHighlight(
        handler: RequestHandler<TextDocumentPositionParams, DocumentHighlight[] | undefined | null, void>
    ): void

    /**
     * Installs a handler for the `DocumentSymbol` request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentSymbol(handler: RequestHandler<DocumentSymbolParams, SymbolInformation[] | undefined | null, void>): void

    /**
     * Installs a handler for the `WorkspaceSymbol` request.
     *
     * @param handler The corresponding handler.
     */
    onWorkspaceSymbol(
        handler: RequestHandler<WorkspaceSymbolParams, SymbolInformation[] | undefined | null, void>
    ): void

    /**
     * Installs a handler for the `CodeAction` request.
     *
     * @param handler The corresponding handler.
     */
    onCodeAction(handler: RequestHandler<CodeActionParams, (Command | CodeAction)[] | undefined | null, void>): void

    /**
     * Compute a list of [lenses](#CodeLens). This call should return as fast as possible and if
     * computing the commands is expensive implementers should only return code lens objects with the
     * range set and handle the resolve request.
     *
     * @param handler The corresponding handler.
     */
    onCodeLens(handler: RequestHandler<CodeLensParams, CodeLens[] | undefined | null, void>): void

    /**
     * This function will be called for each visible code lens, usually when scrolling and after
     * the onCodeLens has been called.
     *
     * @param handler The corresponding handler.
     */
    onCodeLensResolve(handler: RequestHandler<CodeLens, CodeLens, void>): void

    /**
     * Installs a handler for the document formatting request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentFormatting(handler: RequestHandler<DocumentFormattingParams, TextEdit[] | undefined | null, void>): void

    /**
     * Installs a handler for the document range formatting request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentRangeFormatting(
        handler: RequestHandler<DocumentRangeFormattingParams, TextEdit[] | undefined | null, void>
    ): void

    /**
     * Installs a handler for the document on type formatting request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentOnTypeFormatting(
        handler: RequestHandler<DocumentOnTypeFormattingParams, TextEdit[] | undefined | null, void>
    ): void

    /**
     * Installs a handler for the rename request.
     *
     * @param handler The corresponding handler.
     */
    onRenameRequest(handler: RequestHandler<RenameParams, WorkspaceEdit | undefined | null, void>): void

    /**
     * Installs a handler for the document links request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentLinks(handler: RequestHandler<DocumentLinkParams, DocumentLink[] | undefined | null, void>): void

    /**
     * Installs a handler for the document links resolve request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentLinkResolve(handler: RequestHandler<DocumentLink, DocumentLink | undefined | null, void>): void

    /**
     * Installs a handler for the document color request.
     *
     * @param handler The corresponding handler.
     */
    onDocumentColor(handler: RequestHandler<DocumentColorParams, ColorInformation[] | undefined | null, void>): void

    /**
     * Installs a handler for the document color request.
     *
     * @param handler The corresponding handler.
     */
    onColorPresentation(
        handler: RequestHandler<ColorPresentationParams, ColorPresentation[] | undefined | null, void>
    ): void

    /**
     * Installs a handler for the execute command request.
     *
     * @param handler The corresponding handler.
     */
    onExecuteCommand(handler: RequestHandler<ExecuteCommandParams, any | undefined | null, void>): void

    /**
     * Unsubscribes the connection
     */
    unsubscribe(): void
}

export interface IConnection extends Connection {}

export type Feature<B, P> = (Base: new () => B) => new () => B & P

// tslint:disable no-inferred-empty-object-type
export type ConsoleFeature<P> = Feature<RemoteConsole, P>
export function combineConsoleFeatures<O, T>(one: ConsoleFeature<O>, two: ConsoleFeature<T>): ConsoleFeature<O & T> {
    return (Base: new () => RemoteConsole): new () => RemoteConsole & O & T => two(one(Base)) as any
}

export type ContextFeature<P> = Feature<RemoteContext, P>
export function combineContextFeatures<O, T>(one: ContextFeature<O>, two: ContextFeature<T>): ContextFeature<O & T> {
    return (Base: new () => RemoteContext): new () => RemoteContext & O & T => two(one(Base)) as any
}

export type TelemetryFeature<P> = Feature<Telemetry, P>
export function combineTelemetryFeatures<O, T>(
    one: TelemetryFeature<O>,
    two: TelemetryFeature<T>
): TelemetryFeature<O & T> {
    return (Base: new () => Telemetry): new () => Telemetry & O & T => two(one(Base)) as any
}

export type TracerFeature<P> = Feature<Tracer, P>
export function combineTracerFeatures<O, T>(one: TracerFeature<O>, two: TracerFeature<T>): TracerFeature<O & T> {
    return (Base: new () => Tracer): new () => Tracer & O & T => two(one(Base)) as any
}

export type ClientFeature<P> = Feature<RemoteClient, P>
export function combineClientFeatures<O, T>(one: ClientFeature<O>, two: ClientFeature<T>): ClientFeature<O & T> {
    return (Base: new () => RemoteClient): new () => RemoteClient & O & T => two(one(Base)) as any
}
export type WindowFeature<P> = Feature<RemoteWindow, P>
export function combineWindowFeatures<O, T>(one: WindowFeature<O>, two: WindowFeature<T>): WindowFeature<O & T> {
    return (Base: new () => RemoteWindow): new () => RemoteWindow & O & T => two(one(Base)) as any
}
export type WorkspaceFeature<P> = Feature<RemoteWorkspace, P>
export function combineWorkspaceFeatures<O, T>(
    one: WorkspaceFeature<O>,
    two: WorkspaceFeature<T>
): WorkspaceFeature<O & T> {
    return (Base: new () => RemoteWorkspace): new () => RemoteWorkspace & O & T => two(one(Base)) as any
}
// tslint:enable no-inferred-empty-object-type

export interface Features<
    PConsole = _,
    PContext = _,
    PTracer = _,
    PTelemetry = _,
    PClient = _,
    PWindow = _,
    PWorkspace = _
> {
    __brand: 'features'
    console?: ConsoleFeature<PConsole>
    context?: ContextFeature<PContext>
    tracer?: TracerFeature<PTracer>
    telemetry?: TelemetryFeature<PTelemetry>
    client?: ClientFeature<PClient>
    window?: WindowFeature<PWindow>
    workspace?: WorkspaceFeature<PWorkspace>
}
export function combineFeatures<
    OConsole,
    OContext,
    OTracer,
    OTelemetry,
    OClient,
    OWindow,
    OWorkspace,
    TConsole,
    TContext,
    TTracer,
    TTelemetry,
    TClient,
    TWindow,
    TWorkspace
>(
    one: Features<OConsole, OContext, OTracer, OTelemetry, OClient, OWindow, OWorkspace>,
    two: Features<TConsole, TContext, TTracer, TTelemetry, TClient, TWindow, TWorkspace>
): Features<
    OConsole & TConsole,
    OContext & TContext,
    OTracer & TTracer,
    OTelemetry & TTelemetry,
    OClient & TClient,
    OWindow & TWindow,
    OWorkspace & TWorkspace
> {
    function combine<O, T>(one: O | undefined, two: T | undefined, func: (one: O, two: T) => any): any {
        if (one && two) {
            return func(one, two)
        } else if (one) {
            return one
        } else {
            return two
        }
    }
    const result: Features<
        OConsole & TConsole,
        OContext & TContext,
        OTracer & TTracer,
        OTelemetry & TTelemetry,
        OClient & TClient,
        OWindow & TWindow,
        OWorkspace & TWorkspace
    > = {
        __brand: 'features',
        console: combine(one.console, two.console, combineConsoleFeatures),
        context: combine(one.context, two.context, combineContextFeatures),
        tracer: combine(one.tracer, two.tracer, combineTracerFeatures),
        telemetry: combine(one.telemetry, two.telemetry, combineTelemetryFeatures),
        client: combine(one.client, two.client, combineClientFeatures),
        window: combine(one.window, two.window, combineWindowFeatures),
        workspace: combine(one.workspace, two.workspace, combineWorkspaceFeatures),
    }
    return result
}

/**
 * Creates a new connection.
 *
 * @param reader The message reader to read messages from.
 * @param writer The message writer to write message to.
 * @param strategy An optional connection strategy to control additional settings
 * @return a [connection](#IConnection)
 */
export function createConnection<
    PConsole = _,
    PContext = _,
    PTracer = _,
    PTelemetry = _,
    PClient = _,
    PWindow = _,
    PWorkspace = _
>(
    transports: MessageTransports,
    strategy?: ConnectionStrategy,
    factories?: Features<PConsole, PContext, PTracer, PTelemetry, PClient, PWindow, PWorkspace>
): Connection<PConsole, PContext, PTracer, PTelemetry, PClient, PWindow, PWorkspace> {
    // tslint:disable no-inferred-empty-object-type
    const logger = (factories && factories.console
        ? new (factories.console(ConnectionLogger))()
        : new ConnectionLogger()) as ConnectionLogger & PConsole
    const connection = createMessageConnection(transports, logger, strategy)
    logger.rawAttach(connection)
    const context = (factories && factories.context
        ? new (factories.context(RemoteContextImpl))()
        : new RemoteContextImpl()) as RemoteContext & PContext
    const tracer = (factories && factories.tracer
        ? new (factories.tracer(TracerImpl))()
        : new TracerImpl()) as TracerImpl & PTracer
    const telemetry = (factories && factories.telemetry
        ? new (factories.telemetry(TelemetryImpl))()
        : new TelemetryImpl()) as Telemetry & PTelemetry
    const client = (factories && factories.client
        ? new (factories.client(RemoteClientImpl))()
        : new RemoteClientImpl()) as RemoteClient & PClient
    const remoteWindow = (factories && factories.window
        ? new (factories.window(RemoteWindowImpl))()
        : new RemoteWindowImpl()) as RemoteWindow & PWindow
    const workspace = (factories && factories.workspace
        ? new (factories.workspace(RemoteWorkspaceImpl))()
        : new RemoteWorkspaceImpl()) as RemoteWorkspace & PWorkspace
    const allRemotes: Remote[] = [logger, context, tracer, telemetry, client, remoteWindow, workspace]
    // tslint:enable no-inferred-empty-object-type

    let shutdownHandler: RequestHandler<null, void, void> | undefined
    let initializeHandler: RequestHandler<InitializeParams, InitializeResult, InitializeError> | undefined
    let exitHandler: NotificationHandler<null> | undefined
    const protocolConnection: Connection<PConsole, PContext, PTracer, PTelemetry, PClient, PWindow, PWorkspace> = {
        listen: (): void => connection.listen(),

        sendRequest: <R>(type: string | RPCMessageType, ...params: any[]): Promise<R> =>
            connection.sendRequest(typeof type === 'string' ? type : type.method, ...params),
        onRequest: <R, E>(
            type: string | RPCMessageType | StarRequestHandler,
            handler?: GenericRequestHandler<R, E>
        ): void => (connection as any).onRequest(type, handler),

        sendNotification: (type: string | RPCMessageType, param?: any): void => {
            const method = typeof type === 'string' ? type : type.method
            if (param === undefined) {
                connection.sendNotification(method)
            } else {
                connection.sendNotification(method, param)
            }
        },
        onNotification: (
            type: string | RPCMessageType | StarNotificationHandler,
            handler?: GenericNotificationHandler
        ): void => (connection as any).onNotification(type, handler),

        onInitialize: handler => (initializeHandler = handler),
        onInitialized: handler => connection.onNotification(InitializedNotification.type, handler),
        onShutdown: handler => (shutdownHandler = handler),
        onExit: handler => (exitHandler = handler),

        get console(): RemoteConsole & PConsole {
            return logger
        },
        get context(): RemoteContext & PContext {
            return context
        },
        get telemetry(): Telemetry & PTelemetry {
            return telemetry
        },
        get tracer(): Tracer & PTracer {
            return tracer
        },
        get client(): RemoteClient & PClient {
            return client
        },
        get window(): RemoteWindow & PWindow {
            return remoteWindow
        },
        get workspace(): RemoteWorkspace & PWorkspace {
            return workspace
        },

        onDidChangeConfiguration: handler =>
            connection.onNotification(DidChangeConfigurationNotification.type, handler),
        onDidChangeWatchedFiles: handler => connection.onNotification(DidChangeWatchedFilesNotification.type, handler),

        onDidOpenTextDocument: handler => connection.onNotification(DidOpenTextDocumentNotification.type, handler),
        onDidChangeTextDocument: handler => connection.onNotification(DidChangeTextDocumentNotification.type, handler),
        onDidCloseTextDocument: handler => connection.onNotification(DidCloseTextDocumentNotification.type, handler),
        onWillSaveTextDocument: handler => connection.onNotification(WillSaveTextDocumentNotification.type, handler),
        onWillSaveTextDocumentWaitUntil: handler =>
            connection.onRequest(WillSaveTextDocumentWaitUntilRequest.type, handler),
        onDidSaveTextDocument: handler => connection.onNotification(DidSaveTextDocumentNotification.type, handler),

        sendDiagnostics: params => connection.sendNotification(PublishDiagnosticsNotification.type, params),

        onHover: handler => connection.onRequest(HoverRequest.type, handler),
        onCompletion: handler => connection.onRequest(CompletionRequest.type, handler),
        onCompletionResolve: handler => connection.onRequest(CompletionResolveRequest.type, handler),
        onSignatureHelp: handler => connection.onRequest(SignatureHelpRequest.type, handler),
        onDefinition: handler => connection.onRequest(DefinitionRequest.type, handler),
        onTypeDefinition: handler => connection.onRequest(TypeDefinitionRequest.type, handler),
        onImplementation: handler => connection.onRequest(ImplementationRequest.type, handler),
        onReferences: handler => connection.onRequest(ReferencesRequest.type, handler),
        onDocumentHighlight: handler => connection.onRequest(DocumentHighlightRequest.type, handler),
        onDocumentSymbol: handler => connection.onRequest(DocumentSymbolRequest.type, handler),
        onWorkspaceSymbol: handler => connection.onRequest(WorkspaceSymbolRequest.type, handler),
        onCodeAction: handler => connection.onRequest(CodeActionRequest.type, handler),
        onCodeLens: handler => connection.onRequest(CodeLensRequest.type, handler),
        onCodeLensResolve: handler => connection.onRequest(CodeLensResolveRequest.type, handler),
        onDocumentFormatting: handler => connection.onRequest(DocumentFormattingRequest.type, handler),
        onDocumentRangeFormatting: handler => connection.onRequest(DocumentRangeFormattingRequest.type, handler),
        onDocumentOnTypeFormatting: handler => connection.onRequest(DocumentOnTypeFormattingRequest.type, handler),
        onRenameRequest: handler => connection.onRequest(RenameRequest.type, handler),
        onDocumentLinks: handler => connection.onRequest(DocumentLinkRequest.type, handler),
        onDocumentLinkResolve: handler => connection.onRequest(DocumentLinkResolveRequest.type, handler),
        onDocumentColor: handler => connection.onRequest(DocumentColorRequest.type, handler),
        onColorPresentation: handler => connection.onRequest(ColorPresentationRequest.type, handler),
        onExecuteCommand: handler => connection.onRequest(ExecuteCommandRequest.type, handler),

        unsubscribe: () => connection.unsubscribe(),
    }
    for (const remote of allRemotes) {
        remote.attach(protocolConnection)
    }

    connection.onRequest(InitializeRequest.type, params => {
        if (typeof params.trace === 'string') {
            tracer.trace = Trace.fromString(params.trace)
        }
        if (!params.capabilities) {
            params.capabilities = {}
        }
        for (const remote of allRemotes) {
            remote.initialize(params.capabilities)
        }
        if (initializeHandler) {
            const result = initializeHandler(params, new CancellationTokenSource().token)
            return Promise.resolve(result).then(value => {
                if (value instanceof ResponseError) {
                    return value
                }
                let result = value as InitializeResult
                if (!result) {
                    result = { capabilities: {} }
                }
                let capabilities = result.capabilities
                if (!capabilities) {
                    capabilities = {}
                    result.capabilities = capabilities
                }
                for (const remote of allRemotes) {
                    remote.fillServerCapabilities(capabilities)
                }
                return result
            })
        } else {
            const result: InitializeResult = { capabilities: {} }
            for (const remote of allRemotes) {
                remote.fillServerCapabilities(result.capabilities)
            }
            return result
        }
    })

    connection.onRequest<null, void, void, void>(ShutdownRequest.type, () => {
        if (shutdownHandler) {
            return shutdownHandler(null, new CancellationTokenSource().token)
        } else {
            return undefined
        }
    })

    connection.onNotification(ExitNotification.type, () => {
        try {
            if (exitHandler) {
                exitHandler(null)
            }
        } finally {
            protocolConnection.unsubscribe()
        }
    })

    connection.onNotification(SetTraceNotification.type, params => {
        tracer.trace = Trace.fromString(params.value)
    })

    return protocolConnection
}

// Export the protocol currently in proposed state.

export namespace ProposedFeatures {
    export const all: Features<_, _, _, _, _, _> = {
        __brand: 'features',
    }
}

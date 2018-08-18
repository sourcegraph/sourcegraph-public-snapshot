import {
    InitializeParams,
    MessageActionItem,
    MessageType,
    ServerCapabilities,
    ShowInputRequest,
    ShowMessageRequest,
    ShowMessageRequestParams,
} from '../../protocol'
import { IConnection } from '../server'
import { Remote } from './common'

/**
 * The RemoteWindow interface contains all functions to interact with
 * the visual window of VS Code.
 */
export interface RemoteWindow extends Remote {
    /**
     * Show an error message.
     *
     * @param message The message to show.
     */
    showErrorMessage(message: string): void
    showErrorMessage<T extends MessageActionItem>(message: string, ...actions: T[]): Promise<T | undefined>

    /**
     * Show a warning message.
     *
     * @param message The message to show.
     */
    showWarningMessage(message: string): void
    showWarningMessage<T extends MessageActionItem>(message: string, ...actions: T[]): Promise<T | undefined>

    /**
     * Show an information message.
     *
     * @param message The message to show.
     */
    showInformationMessage(message: string): void
    showInformationMessage<T extends MessageActionItem>(message: string, ...actions: T[]): Promise<T | undefined>

    /**
     * Show a request for textual user input.
     *
     * @param message The message to show.
     * @param defaultValue The default value for the user input, or undefined for no default.
     * @returns The user's input, or null if the user (or the client) canceled the input request.
     */
    showInputRequest(message: string, defaultValue?: string): Promise<string | null>
}

export class RemoteWindowImpl implements RemoteWindow {
    private _connection?: IConnection

    public attach(connection: IConnection): void {
        this._connection = connection
    }

    public get connection(): IConnection {
        if (!this._connection) {
            throw new Error('Remote is not attached to a connection yet.')
        }
        return this._connection
    }

    public initialize(_params: InitializeParams): void {
        /* noop */
    }

    public fillServerCapabilities(_capabilities: ServerCapabilities): void {
        /* noop */
    }

    public showErrorMessage(message: string, ...actions: MessageActionItem[]): Promise<MessageActionItem | undefined> {
        const params: ShowMessageRequestParams = { type: MessageType.Error, message, actions }
        return this.connection.sendRequest(ShowMessageRequest.type, params).then(null2Undefined)
    }

    public showWarningMessage(
        message: string,
        ...actions: MessageActionItem[]
    ): Promise<MessageActionItem | undefined> {
        const params: ShowMessageRequestParams = { type: MessageType.Warning, message, actions }
        return this.connection.sendRequest(ShowMessageRequest.type, params).then(null2Undefined)
    }

    public showInformationMessage(
        message: string,
        ...actions: MessageActionItem[]
    ): Promise<MessageActionItem | undefined> {
        const params: ShowMessageRequestParams = { type: MessageType.Info, message, actions }
        return this.connection.sendRequest(ShowMessageRequest.type, params).then(null2Undefined)
    }

    public showInputRequest(message: string, defaultValue?: string): Promise<string | null> {
        return this.connection.sendRequest(ShowInputRequest.type, { message, defaultValue })
    }
}

function null2Undefined<T>(value: T | null): T | undefined {
    if (value === null) {
        return void 0
    }
    return value
}

import { Observable } from 'rxjs';
import * as sourcegraph from 'sourcegraph';
import { MessageActionItem, ShowInputParams, ShowMessageParams, ShowMessageRequestParams } from '../../protocol';
import { Connection } from '../../protocol/jsonrpc2/connection';
import { TextDocumentItem } from '../types/textDocument';
/** @internal */
export interface ClientWindowsAPI {
    $showNotification(message: string): void;
    $showMessage(message: string): Promise<void>;
    $showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined>;
}
/** @internal */
export declare class ClientWindows implements ClientWindowsAPI {
    /** Called when the client receives a window/showMessage notification. */
    private showMessage;
    /**
     * Called when the client receives a window/showMessageRequest request and expected to return a promise
     * that resolves to the selected action.
     */
    private showMessageRequest;
    /**
     * Called when the client receives a window/showInput request and expected to return a promise that
     * resolves to the user's input.
     */
    private showInput;
    private subscriptions;
    private registrations;
    private proxy;
    constructor(connection: Connection, environmentTextDocuments: Observable<TextDocumentItem[] | null>, 
    /** Called when the client receives a window/showMessage notification. */
    showMessage: (params: ShowMessageParams) => void, 
    /**
     * Called when the client receives a window/showMessageRequest request and expected to return a promise
     * that resolves to the selected action.
     */
    showMessageRequest: (params: ShowMessageRequestParams) => Promise<MessageActionItem | null>, 
    /**
     * Called when the client receives a window/showInput request and expected to return a promise that
     * resolves to the user's input.
     */
    showInput: (params: ShowInputParams) => Promise<string | null>);
    $showNotification(message: string): void;
    $showMessage(message: string): Promise<void>;
    $showInputBox(options?: sourcegraph.InputBoxOptions): Promise<string | undefined>;
    unsubscribe(): void;
}

import { Connection } from '../../protocol/jsonrpc2/connection';
import * as plain from '../../protocol/plainTypes';
import { ViewProviderRegistry } from '../providers/view';
/** @internal */
export interface ClientViewsAPI {
    $unregister(id: number): void;
    $registerPanelViewProvider(id: number, provider: {
        id: string;
    }): void;
    $acceptPanelViewUpdate(id: number, params: Partial<plain.PanelView>): void;
}
/** @internal */
export declare class ClientViews implements ClientViewsAPI {
    private viewRegistry;
    private subscriptions;
    private panelViews;
    private registrations;
    constructor(connection: Connection, viewRegistry: ViewProviderRegistry);
    $unregister(id: number): void;
    $registerPanelViewProvider(id: number, provider: {
        id: string;
    }): void;
    $acceptPanelViewUpdate(id: number, params: {
        title?: string;
        content?: string;
    }): void;
    unsubscribe(): void;
}

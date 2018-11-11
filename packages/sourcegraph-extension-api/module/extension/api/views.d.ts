import { Unsubscribable } from 'rxjs';
import * as sourcegraph from 'sourcegraph';
import { ClientViewsAPI } from '../../client/api/views';
/**
 * @internal
 */
declare class ExtPanelView implements sourcegraph.PanelView {
    private proxy;
    private id;
    private subscription;
    private _title;
    private _content;
    constructor(proxy: ClientViewsAPI, id: number, subscription: Unsubscribable);
    title: string;
    content: string;
    unsubscribe(): void;
}
/** @internal */
export interface ExtViewsAPI {
}
/** @internal */
export declare class ExtViews implements ExtViewsAPI {
    private proxy;
    private registrations;
    constructor(proxy: ClientViewsAPI);
    createPanelView(id: string): ExtPanelView;
}
export {};

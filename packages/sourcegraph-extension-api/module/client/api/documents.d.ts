import { Observable } from 'rxjs';
import { TextDocument } from 'sourcegraph';
import { Connection } from '../../protocol/jsonrpc2/connection';
/** @internal */
export declare class ClientDocuments {
    private subscriptions;
    private registrations;
    private proxy;
    constructor(connection: Connection, environmentTextDocuments: Observable<Pick<TextDocument, 'uri' | 'languageId' | 'text'>[] | null>);
    unsubscribe(): void;
}

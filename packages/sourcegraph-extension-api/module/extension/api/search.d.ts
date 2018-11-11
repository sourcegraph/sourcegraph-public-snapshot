import { Unsubscribable } from 'rxjs';
import { QueryTransformer } from 'sourcegraph';
import { SearchAPI } from 'src/client/api/search';
export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>;
}
export declare class ExtSearch implements ExtSearchAPI {
    private proxy;
    private registrations;
    constructor(proxy: SearchAPI);
    registerQueryTransformer(provider: QueryTransformer): Unsubscribable;
    $transformQuery(id: number, query: string): Promise<string>;
}

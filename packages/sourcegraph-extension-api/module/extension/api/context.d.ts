import { ContextValues } from 'sourcegraph';
import { ClientContextAPI } from '../../client/api/context';
/** @internal */
export interface ExtContextAPI {
}
/** @internal */
export declare class ExtContext implements ExtContextAPI {
    private proxy;
    constructor(proxy: ClientContextAPI);
    updateContext(updates: ContextValues): void;
}

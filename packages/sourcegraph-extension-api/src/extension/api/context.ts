import { ContextValues } from 'sourcegraph'
import { ClientContextAPI } from '../../client/api/context'

/** @internal */
export interface ExtContextAPI {}

/** @internal */
export class ExtContext implements ExtContextAPI {
    constructor(private proxy: ClientContextAPI) {}

    public updateContext(updates: ContextValues): void {
        this.proxy.$acceptContextUpdates(updates)
    }
}

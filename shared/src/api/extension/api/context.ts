import { ContextValues } from 'sourcegraph'
import { ClientContextAPI } from '../../client/api/context'

/** @internal */
export class ExtContext {
    constructor(private proxy: ClientContextAPI) {}

    public updateContext(updates: ContextValues): void {
        this.proxy.$acceptContextUpdates(updates)
    }
}

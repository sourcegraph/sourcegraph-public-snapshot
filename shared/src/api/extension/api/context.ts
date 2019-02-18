import { ProxyResult } from '@sourcegraph/comlink'
import { ContextValues } from 'sourcegraph'
import { ClientContextAPI } from '../../client/api/context'

/** @internal */
export class ExtContext {
    constructor(private proxy: ProxyResult<ClientContextAPI>) {}

    public updateContext(updates: ContextValues): void {
        // tslint:disable-next-line: no-floating-promises
        this.proxy.$acceptContextUpdates(updates)
    }
}

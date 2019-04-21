import * as comlink from '@sourcegraph/comlink'
import { Observable } from 'rxjs'
import { createBarrier } from '../../integration-test/testHelpers'
import { proxySubscribable } from './common'

describe('proxySubscribable', () => {
    test('unsubscribes', async () => {
        const gotUnsubscribed = createBarrier()
        const observable = new Observable<number>(sub => {
            sub.next(1)
            return () => gotUnsubscribed.done()
        })

        const gotValue = createBarrier()
        proxySubscribable(observable)
            .subscribe({
                [comlink.proxyMarker]: Promise.resolve(true),
                next: async value => {
                    expect(value).toBe(1)
                    gotValue.done()
                },
                error: () => {
                    throw new Error('error')
                },
                complete: () => {
                    throw new Error('unexpected complete')
                },
            })
            .unsubscribe()
        await gotUnsubscribed.wait
        await gotValue.wait
    })
})

import * as comlink from '@sourcegraph/comlink'
import { from, Observable, Subscribable } from 'rxjs'
import { first, toArray } from 'rxjs/operators'
import { createBarrier } from '../api/integration-test/testHelpers'

describe('transferHandlers', () => {
    test('Observable', async () => {
        let unsubscribed = 0
        let subscribed = 0
        const gotUnsubscribed = createBarrier()
        const observable = new Observable<number>(sub => {
            subscribed++
            sub.next(1)
            return () => {
                unsubscribed++
                gotUnsubscribed.done()
            }
        })

        const { port1, port2 } = new MessageChannel()
        comlink.expose(() => observable, port1)
        const getObservable = comlink.wrap<() => Subscribable<number>>(port2)

        expect(
            await from(getObservable())
                .pipe(
                    first(),
                    toArray()
                )
                .toPromise()
        ).toEqual([1])
        await gotUnsubscribed.wait
        expect(unsubscribed).toBe(1)
        expect(subscribed).toBe(1)
    })
})

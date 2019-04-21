import * as comlink from '@sourcegraph/comlink'
import { from, Observable } from 'rxjs'
import { first, take } from 'rxjs/operators'
import { ProxySubscribable } from '../../extension/api/common'
import { createBarrier } from '../../integration-test/testHelpers'
import { wrapRemoteObservable } from './common'

describe('wrapRemoteObservable', () => {
    describe('unsubscribes', () => {
        test('when using without comlink', done => {
            const observable = new Observable<number>(() => () => done())
            const proxyObservable = comlink.proxy(() => observable)
            const remoteObservable = wrapRemoteObservable<number>(proxyObservable())
            remoteObservable.subscribe().unsubscribe()
        })

        test('when using comlink', async () => {
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

            const wrapper = new MessageChannel()
            comlink.expose(() => observable, wrapper.port1)

            const remoteGetObservable = comlink.wrap<() => ProxySubscribable<number>>(wrapper.port2)
            const getObservable = () => wrapRemoteObservable<number>(remoteGetObservable())

            const sub = getObservable()
                .pipe(first()) // TODO!(sqs): first shouldnt be necessary
                .subscribe()
            await new Promise(resolve => setTimeout(resolve))
            sub.unsubscribe()

            await gotUnsubscribed.wait
            expect(unsubscribed).toBe(1)
            expect(subscribed).toBe(1)
        })
    })
})

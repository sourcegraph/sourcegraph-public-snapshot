import * as comlink from '@sourcegraph/comlink'
import { Observable } from 'rxjs'
import { wrapRemoteObservable } from './common'

describe('wrapRemoteObservable', () => {
    test('unsubscribes', done => {
        const observable = new Observable<number>(() => () => done())
        const proxyObservable = comlink.proxy(() => observable)
        const remoteObservable = wrapRemoteObservable<number>(proxyObservable())
        remoteObservable.subscribe().unsubscribe()
    })
})

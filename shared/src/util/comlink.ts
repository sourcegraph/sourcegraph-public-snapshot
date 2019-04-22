import { proxyMarker, transferHandlers } from '@sourcegraph/comlink'
import { isObservable, observable } from 'rxjs'
import { wrapRemoteObservable } from '../api/client/api/common'
import { proxySubscribable } from '../api/extension/api/common'

// transferHandlers.set('Unsubscribable', {
//     canHandle: obj => 'unsubscribe' in obj && typeof obj.unsubscribe === 'function',
//     serialize: ev => [
//         {
//             type: ev.type,
//             data: ev.data,
//         },
//         [],
//     ],
//     deserialize: obj => obj,
// })

const proxyTransferHandler = transferHandlers.get('proxy')!

transferHandlers.set('Observable', {
    canHandle: obj => isObservable(obj) && !obj[proxyMarker],
    serialize: observable => {
        const obj = proxySubscribable(observable)
        return proxyTransferHandler.serialize(obj)
    },
    deserialize: obj => wrapRemoteObservable(proxyTransferHandler.deserialize(obj)),
})

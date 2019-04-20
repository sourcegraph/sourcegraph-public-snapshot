import { transferHandlers } from '@sourcegraph/comlink'

transferHandlers.set('MessageEvent', {
    canHandle: obj => obj instanceof Event && obj.type === 'message',
    serialize: ev => [
        {
            type: ev.type,
            data: ev.data,
        },
        [],
    ],
    deserialize: obj => obj,
})

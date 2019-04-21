import { transferHandlers } from '@sourcegraph/comlink'

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

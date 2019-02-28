import H from 'history'
import { Subject } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { ActivationStatus, percentageDone } from './Activation'

describe('ActivationStatus', () => {
    test('activate', async () => {
        let underlyingActivationStatus = {}
        const fetchReturned = new Subject() // used to trigger returns of fetches
        const activation = new ActivationStatus(
            [
                {
                    id: 'id1',
                    title: 'title1',
                    detail: 'detail1',
                    action: (h: H.History) => void 0,
                },
                {
                    id: 'id2',
                    title: 'title2',
                    detail: 'detail2',
                    action: (h: H.History) => void 0,
                },
                {
                    id: 'id3',
                    title: 'title3',
                    detail: 'detail3',
                    action: (h: H.History) => void 0,
                },
            ],
            () => fetchReturned.pipe(map(() => underlyingActivationStatus))
        )

        const update = new Promise(resolve => activation.updateCompleted.pipe(first()).subscribe(u => resolve(u)))
        activation.update({ id2: true })

        // Initial fetch not yet returned
        expect(activation.completed.value).toEqual(null)

        // Return initial fetch
        fetchReturned.next()
        expect(activation.completed.value).toEqual({ id1: false, id2: false, id3: false })

        // Correct update after initial fetch
        expect(await update).toEqual({ id2: true })
        // Correct state after update
        expect(activation.completed.value).toEqual({ id1: false, id2: true, id3: false })

        // Refetch and check correct activation state
        underlyingActivationStatus = { id1: true, id2: false, id3: false }
        activation.refetch()
        fetchReturned.next()
        expect(activation.completed.value).toEqual({ id1: true, id2: false, id3: false })

        // Update
        const update2 = new Promise(resolve => activation.updateCompleted.pipe(first()).subscribe(u => resolve(u)))
        activation.update({ id3: true })
        expect(await update2).toEqual({ id3: true })
        expect(activation.completed.value).toEqual({ id1: true, id2: false, id3: true })
    })
    test('percentageDone', () => {
        expect(percentageDone({ id1: true })).toEqual(100)
        expect(percentageDone({ id1: false })).toEqual(0)
        expect(percentageDone({ id1: false, id2: true })).toEqual(50)
    })
})

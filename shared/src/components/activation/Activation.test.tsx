import H from 'history'
import { of } from 'rxjs'
import { ActivationStatus, percentageDone } from './Activation'

describe('Activation', () => {
    test('activate', () => {
        let underlyingActivationStatus = {}
        const activation = new ActivationStatus(
            [
                {
                    id: 'id1',
                    title: 'title1',
                    detail: 'detail1',
                    action: (h: H.History) => void 0,
                },
            ],
            () => of(underlyingActivationStatus)
        )

        expect(activation.completed.value).toEqual(null)

        activation.update(null)
        expect(activation.completed.value).toEqual({ id1: false })

        underlyingActivationStatus = { id1: true }
        activation.update(null)
        expect(activation.completed.value).toEqual({ id1: true })
    })
    test('percentageDone', () => {
        expect(percentageDone({ id1: true })).toEqual(100)
        expect(percentageDone({ id1: false })).toEqual(0)
        expect(percentageDone({ id1: false, id2: true })).toEqual(50)
    })
})

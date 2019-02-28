import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'
import { ActivationStatus } from './Activation'
import { ActivationDropdown } from './ActivationDropdown'

describe('ActivationDropdown', () => {
    test('render 0/2', () => {
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
            ],
            () => of({})
        )
        const history = H.createMemoryHistory({ keyLength: 0 })
        const component = renderer.create(<ActivationDropdown history={history} activation={activation} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('render 1/2', () => {
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
            ],
            () => of({ id1: true })
        )
        const history = H.createMemoryHistory({ keyLength: 0 })
        const component = renderer.create(<ActivationDropdown history={history} activation={activation} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
})

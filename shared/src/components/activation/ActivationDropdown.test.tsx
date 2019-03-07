import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { noop } from 'rxjs'
import { setLinkComponent } from '../../components/Link'
import { Activation } from './Activation'
import { ActivationDropdown } from './ActivationDropdown'

describe('ActivationDropdown', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    const baseActivation: Activation = {
        steps: [
            {
                id: 'ConnectedCodeHost',
                title: 'title1',
                detail: 'detail1',
                onClick: noop,
            },
            {
                id: 'EnabledRepository',
                title: 'title2',
                detail: 'detail2',
                link: { to: '/some/url' },
            },
        ],
        refetch: noop,
        update: noop,
    }
    const history = H.createMemoryHistory({ keyLength: 0 })
    test('render loading', () => {
        const component = renderer.create(<ActivationDropdown history={history} activation={baseActivation} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('render 0/2 completed', () => {
        const component = renderer.create(
            <ActivationDropdown history={history} activation={{ ...baseActivation, completed: {} }} />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('render 1/2 completed', () => {
        const component = renderer.create(
            <ActivationDropdown
                history={history}
                activation={{ ...baseActivation, completed: { ConnectedCodeHost: true } }}
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})

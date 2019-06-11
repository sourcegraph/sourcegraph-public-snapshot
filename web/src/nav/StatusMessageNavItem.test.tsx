import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from '../../../shared/src/components/Link'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('UserNavItem', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    test('nothing cloning', () => {
        expect(renderer.create(<StatusMessagesNavItem />).toJSON()).toMatchSnapshot()
    })
})

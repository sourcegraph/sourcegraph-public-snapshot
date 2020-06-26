import React from 'react'
import { ProductSubscriptionUserCountFormControl } from './ProductSubscriptionUserCountFormControl'
import { mount } from 'enzyme'

describe('ProductSubscriptionUserCountFormControl', () => {
    test('renders', () => {
        expect(
            mount(<ProductSubscriptionUserCountFormControl value={123} onChange={() => undefined} />).children()
        ).toMatchSnapshot()
    })
})

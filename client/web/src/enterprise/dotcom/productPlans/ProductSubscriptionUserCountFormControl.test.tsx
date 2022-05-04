import { render } from '@testing-library/react'

import { ProductSubscriptionUserCountFormControl } from './ProductSubscriptionUserCountFormControl'

describe('ProductSubscriptionUserCountFormControl', () => {
    test('renders', () => {
        expect(
            render(<ProductSubscriptionUserCountFormControl value={123} onChange={() => undefined} />).asFragment()
        ).toMatchSnapshot()
    })
})

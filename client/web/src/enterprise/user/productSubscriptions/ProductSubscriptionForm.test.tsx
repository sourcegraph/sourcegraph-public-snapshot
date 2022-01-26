import { render } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import React from 'react'
import { Router } from 'react-router'

import { ProductSubscriptionForm } from './ProductSubscriptionForm'

jest.mock('../../dotcom/billing/StripeWrapper', () => ({
    StripeWrapper: ({
        component: Component,
        ...props
    }: {
        component: React.ComponentType<{ stripe: unknown }>
        [name: string]: unknown
    }) => <Component {...props} stripe={{}} />,
}))

jest.mock('react-stripe-elements', () => ({ CardElement: 'CardElement' }))

describe('ProductSubscriptionForm', () => {
    test('new subscription for anonymous viewer (no account)', () => {
        const history = createMemoryHistory()
        expect(
            render(
                <Router history={history}>
                    <ProductSubscriptionForm
                        accountID={null}
                        subscriptionID={null}
                        onSubmit={() => undefined}
                        submissionState={undefined}
                        primaryButtonText="Submit"
                        isLightTheme={false}
                        history={history}
                    />
                </Router>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('new subscription for existing account', () => {
        const history = createMemoryHistory()
        expect(
            render(
                <Router history={history}>
                    <ProductSubscriptionForm
                        accountID="a"
                        subscriptionID={null}
                        onSubmit={() => undefined}
                        submissionState={undefined}
                        primaryButtonText="Submit"
                        isLightTheme={false}
                        history={history}
                    />
                </Router>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('edit existing subscription', () => {
        const history = createMemoryHistory()
        expect(
            render(
                <Router history={history}>
                    <ProductSubscriptionForm
                        accountID="a"
                        subscriptionID="s"
                        initialValue={{ userCount: 123, billingPlanID: 'p' }}
                        onSubmit={() => undefined}
                        submissionState={undefined}
                        primaryButtonText="Submit"
                        isLightTheme={false}
                        history={history}
                    />
                </Router>
            ).asFragment()
        ).toMatchSnapshot()
    })
})

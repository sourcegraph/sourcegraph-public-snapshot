import React from 'react'

import { createMemoryHistory } from 'history'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { ProductSubscriptionForm } from './ProductSubscriptionForm'

jest.mock('../../dotcom/billing/StripeWrapper', () => ({
    StripeWrapper: ({
        component: Component,
        ...props
    }: {
        component: React.ComponentType<React.PropsWithChildren<{ stripe: unknown }>>
        [name: string]: unknown
    }) => <Component {...props} stripe={{}} />,
}))

jest.mock('react-stripe-elements', () => ({ CardElement: 'CardElement' }))

describe('ProductSubscriptionForm', () => {
    test('new subscription for anonymous viewer (no account)', () => {
        const history = createMemoryHistory()
        expect(
            renderWithBrandedContext(
                <ProductSubscriptionForm
                    accountID={null}
                    subscriptionID={null}
                    onSubmit={() => undefined}
                    submissionState={undefined}
                    primaryButtonText="Submit"
                    isLightTheme={false}
                    history={history}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('new subscription for existing account', () => {
        const history = createMemoryHistory()
        expect(
            renderWithBrandedContext(
                <ProductSubscriptionForm
                    accountID="a"
                    subscriptionID={null}
                    onSubmit={() => undefined}
                    submissionState={undefined}
                    primaryButtonText="Submit"
                    isLightTheme={false}
                    history={history}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('edit existing subscription', () => {
        const history = createMemoryHistory()
        expect(
            renderWithBrandedContext(
                <ProductSubscriptionForm
                    accountID="a"
                    subscriptionID="s"
                    initialValue={{ userCount: 123, billingPlanID: 'p' }}
                    onSubmit={() => undefined}
                    submissionState={undefined}
                    primaryButtonText="Submit"
                    isLightTheme={false}
                    history={history}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })
})

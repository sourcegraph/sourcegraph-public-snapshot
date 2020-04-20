import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer from 'react-test-renderer'
import { ProductSubscriptionForm } from './ProductSubscriptionForm'

jest.mock('../../dotcom/billing/StripeWrapper', () => ({
    StripeWrapper: ({
        component: C,
        ...props
    }: {
        component: React.ComponentType<{ stripe: unknown }>
        [name: string]: unknown
    }) => <C {...props} stripe={{}} />,
}))

jest.mock('react-stripe-elements', () => ({ CardElement: 'CardElement' }))

jest.mock('../../dotcom/productPlans/ProductSubscriptionUserCountFormControl', () => ({
    ProductSubscriptionUserCountFormControl: 'ProductSubscriptionUserCountFormControl',
}))

jest.mock('../../dotcom/productPlans/ProductPlanFormControl', () => ({
    ProductPlanFormControl: 'ProductPlanFormControl',
}))

jest.mock('./NewProductSubscriptionPaymentSection', () => ({
    NewProductSubscriptionPaymentSection: 'NewProductSubscriptionPaymentSection',
}))

describe('ProductSubscriptionForm', () => {
    test('new subscription for anonymous viewer (no account)', () => {
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <ProductSubscriptionForm
                            accountID={null}
                            subscriptionID={null}
                            onSubmit={() => undefined}
                            submissionState={undefined}
                            primaryButtonText="Submit"
                            isLightTheme={false}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('new subscription for existing account', () => {
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <ProductSubscriptionForm
                            accountID="a"
                            subscriptionID={null}
                            onSubmit={() => undefined}
                            submissionState={undefined}
                            primaryButtonText="Submit"
                            isLightTheme={false}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    test('edit existing subscription', () => {
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <ProductSubscriptionForm
                            accountID="a"
                            subscriptionID="s"
                            initialValue={{ userCount: 123, billingPlanID: 'p' }}
                            onSubmit={() => undefined}
                            submissionState={undefined}
                            primaryButtonText="Submit"
                            isLightTheme={false}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})

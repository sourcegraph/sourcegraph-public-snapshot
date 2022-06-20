import { screen, waitFor } from '@testing-library/react'
import { createMemoryHistory } from 'history'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { ProductSubscriptionForm } from './ProductSubscriptionForm'

jest.mock('../../dotcom/productSubscriptions/features', () => ({
    billingPublishableKey: 'publishable-key',
}))

jest.mock('@stripe/stripe-js', () => ({
    ...jest.requireActual('@stripe/stripe-js'),
    loadStripe: () =>
        Promise.resolve({
            elements: () => {},
            createToken: () => {},
            createPaymentMethod: () => {},
            confirmCardPayment: () => {},
        }),
}))

jest.mock('@stripe/react-stripe-js', () => ({
    ...jest.requireActual('@stripe/react-stripe-js'),
    CardElement: 'cardelement',
}))

jest.mock('../../dotcom/productPlans/ProductPlanFormControl', () => ({
    ProductPlanFormControl: 'productplanformcontrol',
}))

jest.mock('../../user/productSubscriptions/NewProductSubscriptionPaymentSection.tsx', () => ({
    NewProductSubscriptionPaymentSection: 'newproductsubscriptionpaymentsection',
}))

describe('ProductSubscriptionForm', () => {
    test('new subscription for anonymous viewer (no account)', async () => {
        const history = createMemoryHistory()
        const { asFragment } = renderWithBrandedContext(
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
        )
        await waitFor(() => expect(screen.getByText(/submit/i)).toBeInTheDocument())
        expect(asFragment()).toMatchSnapshot()
    })

    test('new subscription for existing account', async () => {
        const history = createMemoryHistory()
        const { asFragment } = renderWithBrandedContext(
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
        )
        await waitFor(() => expect(screen.getByText(/submit/i)).toBeInTheDocument())
        expect(asFragment()).toMatchSnapshot()
    })

    test('edit existing subscription', async () => {
        const history = createMemoryHistory()
        const { asFragment } = renderWithBrandedContext(
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
        )
        await waitFor(() => expect(screen.getByText(/submit/i)).toBeInTheDocument())
        expect(asFragment()).toMatchSnapshot()
    })
})

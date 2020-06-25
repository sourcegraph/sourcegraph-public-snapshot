import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { NewProductSubscriptionPaymentSection } from './NewProductSubscriptionPaymentSection'
import { of } from 'rxjs'
import { mount } from 'enzyme'

jest.mock('./ProductSubscriptionBeforeAfterInvoiceItem', () => ({
    ProductSubscriptionBeforeAfterInvoiceItem: 'ProductSubscriptionBeforeAfterInvoiceItem',
}))

describe('NewProductSubscriptionPaymentSection', () => {
    test('new subscription', () => {
        expect(
            mount(
                <NewProductSubscriptionPaymentSection
                    accountID={null}
                    subscriptionID={null}
                    productSubscription={{ billingPlanID: 'p', userCount: 123 }}
                    onValidityChange={() => undefined}
                    _queryPreviewProductSubscriptionInvoice={() =>
                        of<GQL.IProductSubscriptionPreviewInvoice>({
                            __typename: 'ProductSubscriptionPreviewInvoice',
                            beforeInvoiceItem: null,
                            afterInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2020-01-01',
                                userCount: 123,
                                plan: {} as GQL.IProductPlan,
                            },
                            prorationDate: null,
                            isDowngradeRequiringManualIntervention: false,
                            price: 12345,
                        })
                    }
                />
            )
        ).toMatchSnapshot()
    })

    test('valid change to existing subscription', () => {
        expect(
            mount(
                <NewProductSubscriptionPaymentSection
                    accountID="a"
                    subscriptionID="s"
                    productSubscription={{ billingPlanID: 'p', userCount: 123 }}
                    onValidityChange={() => undefined}
                    _queryPreviewProductSubscriptionInvoice={() =>
                        of<GQL.IProductSubscriptionPreviewInvoice>({
                            __typename: 'ProductSubscriptionPreviewInvoice',
                            beforeInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2020-01-01',
                                userCount: 123,
                                plan: {} as GQL.IProductPlan,
                            },
                            afterInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2021-01-01',
                                userCount: 456,
                                plan: {} as GQL.IProductPlan,
                            },
                            prorationDate: null,
                            isDowngradeRequiringManualIntervention: false,
                            price: 23456,
                        })
                    }
                />
            )
        ).toMatchSnapshot()
    })

    test('no change to existing subscription', () => {
        expect(
            mount(
                <NewProductSubscriptionPaymentSection
                    accountID="a"
                    subscriptionID="s"
                    productSubscription={{ billingPlanID: 'p', userCount: 123 }}
                    onValidityChange={() => undefined}
                    _queryPreviewProductSubscriptionInvoice={() =>
                        of<GQL.IProductSubscriptionPreviewInvoice>({
                            __typename: 'ProductSubscriptionPreviewInvoice',
                            beforeInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2020-01-01',
                                userCount: 123,
                                plan: {} as GQL.IProductPlan,
                            },
                            afterInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2020-01-01',
                                userCount: 123,
                                plan: {} as GQL.IProductPlan,
                            },
                            prorationDate: null,
                            isDowngradeRequiringManualIntervention: false,
                            price: 0,
                        })
                    }
                />
            )
        ).toMatchSnapshot()
    })

    test('downgrade to existing subscription', () => {
        expect(
            mount(
                <NewProductSubscriptionPaymentSection
                    accountID="a"
                    subscriptionID="s"
                    productSubscription={{ billingPlanID: 'p', userCount: 123 }}
                    onValidityChange={() => undefined}
                    _queryPreviewProductSubscriptionInvoice={() =>
                        of<GQL.IProductSubscriptionPreviewInvoice>({
                            __typename: 'ProductSubscriptionPreviewInvoice',
                            beforeInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2020-01-01',
                                userCount: 123,
                                plan: {} as GQL.IProductPlan,
                            },
                            afterInvoiceItem: {
                                __typename: 'ProductSubscriptionInvoiceItem',
                                expiresAt: '2021-01-01',
                                userCount: 456,
                                plan: {} as GQL.IProductPlan,
                            },
                            prorationDate: null,
                            isDowngradeRequiringManualIntervention: true,
                            price: 23456,
                        })
                    }
                />
            )
        ).toMatchSnapshot()
    })
})

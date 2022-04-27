import { render, act } from '@testing-library/react'
import { of } from 'rxjs'

import * as GQL from '@sourcegraph/shared/src/schema'

import { ProductPlanFormControl } from './ProductPlanFormControl'

describe('ProductPlanFormControl', () => {
    test('new subscription', () => {
        const component = render(
            <ProductPlanFormControl
                value="p"
                onChange={() => undefined}
                _queryProductPlans={() =>
                    of<GQL.IProductPlan[]>([
                        {
                            __typename: 'ProductPlan',
                            billingPlanID: 'p0',
                            productPlanID: 'pp0',
                            minQuantity: null,
                            maxQuantity: 10,
                            name: 'n0',
                            nameWithBrand: 'nb0',
                            planTiers: [],
                            pricePerUserPerYear: 123,
                            tiersMode: 'volume',
                        },
                        {
                            __typename: 'ProductPlan',
                            billingPlanID: 'p1',
                            productPlanID: 'pp1',
                            minQuantity: null,
                            maxQuantity: null,
                            name: 'n1',
                            nameWithBrand: 'nb1',
                            planTiers: [],
                            pricePerUserPerYear: 456,
                            tiersMode: 'volume',
                        },
                    ])
                }
            />
        )
        act(() => undefined)
        expect(component.asFragment()).toMatchSnapshot()
    })
})

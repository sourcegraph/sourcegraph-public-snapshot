import { render } from '@testing-library/react'

import { ProductPlanPrice } from './ProductPlanPrice'

describe('ProductPlanPrice', () => {
    test('free, no max', () => {
        expect(
            render(
                <ProductPlanPrice
                    plan={{
                        minQuantity: null,
                        maxQuantity: null,
                        tiersMode: '',
                        pricePerUserPerYear: 0,
                        planTiers: [],
                    }}
                />
            ).asFragment()
        ).toMatchInlineSnapshot(`
            <DocumentFragment>
              $0/month 
            </DocumentFragment>
        `)
    })

    test('max', () => {
        expect(
            render(
                <ProductPlanPrice
                    plan={{
                        minQuantity: 50,
                        maxQuantity: 100,
                        tiersMode: '',
                        pricePerUserPerYear: 0,
                        planTiers: [],
                    }}
                />
            ).asFragment()
        ).toMatchInlineSnapshot(`
            <DocumentFragment>
              $0/month (up to 100 users)
            </DocumentFragment>
        `)
    })

    test('priced', () => {
        expect(
            render(
                <ProductPlanPrice
                    plan={{
                        minQuantity: null,
                        maxQuantity: null,
                        tiersMode: '',
                        pricePerUserPerYear: 100,
                        planTiers: [],
                    }}
                />
            ).asFragment()
        ).toMatchInlineSnapshot(`
            <DocumentFragment>
              $0.08/user/month (paid yearly)
            </DocumentFragment>
        `)
    })
})

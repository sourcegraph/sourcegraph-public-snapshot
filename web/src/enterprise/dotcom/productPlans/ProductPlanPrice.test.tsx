import React from 'react'
import { ProductPlanPrice } from './ProductPlanPrice'
import { mount } from 'enzyme'

describe('ProductPlanPrice', () => {
    test('free, no max', () => {
        expect(
            mount(
                <ProductPlanPrice
                    plan={{
                        minQuantity: null,
                        maxQuantity: null,
                        tiersMode: '',
                        pricePerUserPerYear: 0,
                        planTiers: [],
                    }}
                />
            )
        ).toMatchInlineSnapshot(`
            <ProductPlanPrice
              plan={
                Object {
                  "maxQuantity": null,
                  "minQuantity": null,
                  "planTiers": Array [],
                  "pricePerUserPerYear": 0,
                  "tiersMode": "",
                }
              }
            >
              $0/month
               
            </ProductPlanPrice>
        `)
    })

    test('max', () => {
        expect(
            mount(
                <ProductPlanPrice
                    plan={{
                        minQuantity: 50,
                        maxQuantity: 100,
                        tiersMode: '',
                        pricePerUserPerYear: 0,
                        planTiers: [],
                    }}
                />
            )
        ).toMatchInlineSnapshot(`
            <ProductPlanPrice
              plan={
                Object {
                  "maxQuantity": 100,
                  "minQuantity": 50,
                  "planTiers": Array [],
                  "pricePerUserPerYear": 0,
                  "tiersMode": "",
                }
              }
            >
              $0/month
               
              (up to 
              100
               
              users
              )
            </ProductPlanPrice>
        `)
    })

    test('priced', () => {
        expect(
            mount(
                <ProductPlanPrice
                    plan={{
                        minQuantity: null,
                        maxQuantity: null,
                        tiersMode: '',
                        pricePerUserPerYear: 100,
                        planTiers: [],
                    }}
                />
            )
        ).toMatchInlineSnapshot(`
            <ProductPlanPrice
              plan={
                Object {
                  "maxQuantity": null,
                  "minQuantity": null,
                  "planTiers": Array [],
                  "pricePerUserPerYear": 100,
                  "tiersMode": "",
                }
              }
            >
              $0.08
              /user/month (paid yearly)
            </ProductPlanPrice>
        `)
    })
})

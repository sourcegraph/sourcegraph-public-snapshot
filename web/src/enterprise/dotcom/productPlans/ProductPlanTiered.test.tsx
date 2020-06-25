import React from 'react'
import { ProductPlanTiered } from './ProductPlanTiered'
import { mount } from 'enzyme'

describe('ProductPlanTiered', () => {
    test('volume', () => {
        expect(
            mount(
                <ProductPlanTiered
                    plan={{
                        minQuantity: null,
                        tiersMode: 'volume',
                        planTiers: [
                            { __typename: 'PlanTier', flatAmount: 100, unitAmount: 200, upTo: 300 },
                            { __typename: 'PlanTier', flatAmount: 400, unitAmount: 500, upTo: 600 },
                        ],
                    }}
                />
            )
        ).toMatchInlineSnapshot(`
            <ProductPlanTiered
              plan={
                Object {
                  "minQuantity": null,
                  "planTiers": Array [
                    Object {
                      "__typename": "PlanTier",
                      "flatAmount": 100,
                      "unitAmount": 200,
                      "upTo": 300,
                    },
                    Object {
                      "__typename": "PlanTier",
                      "flatAmount": 400,
                      "unitAmount": 500,
                      "upTo": 600,
                    },
                  ],
                  "tiersMode": "volume",
                }
              }
            >
              <div
                key="0"
              >
                $0.17/user/month
                 
                for 1–300 users
              </div>
              <div
                key="1"
              >
                $0.42/user/month
                 
                for 301–600 users
              </div>
            </ProductPlanTiered>
        `)
    })

    test('volume', () => {
        expect(
            mount(
                <ProductPlanTiered
                    plan={{
                        minQuantity: 50,
                        tiersMode: 'graduated',
                        planTiers: [
                            { __typename: 'PlanTier', flatAmount: 100, unitAmount: 200, upTo: 300 },
                            { __typename: 'PlanTier', flatAmount: 400, unitAmount: 500, upTo: 600 },
                        ],
                    }}
                />
            )
        ).toMatchInlineSnapshot(`
            <ProductPlanTiered
              plan={
                Object {
                  "minQuantity": 50,
                  "planTiers": Array [
                    Object {
                      "__typename": "PlanTier",
                      "flatAmount": 100,
                      "unitAmount": 200,
                      "upTo": 300,
                    },
                    Object {
                      "__typename": "PlanTier",
                      "flatAmount": 400,
                      "unitAmount": 500,
                      "upTo": 600,
                    },
                  ],
                  "tiersMode": "graduated",
                }
              }
            >
              <div
                key="0"
              >
                $0.17/user/month
                 
                for the first 300 users
              </div>
              <div
                key="1"
              >
                $0.42/user/month
                 
                for the next 300 users
              </div>
            </ProductPlanTiered>
        `)
    })
})

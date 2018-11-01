import React from 'react'
import * as GQL from '../../../../packages/webapp/src/backend/graphqlschema'

/** Displays the price of a plan. */
export const ProductPlanPrice: React.SFC<{
    pricePerUserPerYear: GQL.IProductPlan['pricePerUserPerYear']
}> = ({ pricePerUserPerYear }) => (
    <>
        {(pricePerUserPerYear / 100 /* cents in a USD */ / 12) /* months */
            .toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 })}
        /user/month (paid yearly)
    </>
)

import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { numberWithCommas } from '@sourcegraph/webapp/dist/util/strings'
import React from 'react'

/** Displays the price of a plan. */
export const ProductPlanPrice: React.SFC<{
    pricePerUserPerYear: GQL.IProductPlan['pricePerUserPerYear']
}> = ({ pricePerUserPerYear }) => (
    <>
        ${numberWithCommas(pricePerUserPerYear / 100 /* cents in a USD */ / 12 /* months */)}
        /user/month (paid yearly)
    </>
)

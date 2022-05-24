import React from 'react'

import { pluralize } from '@sourcegraph/common'
import * as GQL from '@sourcegraph/shared/src/schema'

import { ProductPlanTiered } from './ProductPlanTiered'

/** Displays the price of a plan. */
export const ProductPlanPrice: React.FunctionComponent<
    React.PropsWithChildren<{
        plan: Pick<GQL.IProductPlan, 'pricePerUserPerYear' | 'planTiers' | 'tiersMode' | 'minQuantity' | 'maxQuantity'>
    }>
> = ({ plan }) =>
    plan.planTiers.length > 0 ? (
        <ProductPlanTiered plan={plan} />
    ) : plan.pricePerUserPerYear === 0 ? (
        <>
            $0/month{' '}
            {plan.maxQuantity !== null && (
                <>
                    (up to {plan.maxQuantity} {pluralize('user', plan.maxQuantity)})
                </>
            )}
        </>
    ) : (
        <>
            {(plan.pricePerUserPerYear / 100 /* cents in a USD */ / 12) /* months */
                .toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0 })}
            /user/month (paid yearly)
        </>
    )

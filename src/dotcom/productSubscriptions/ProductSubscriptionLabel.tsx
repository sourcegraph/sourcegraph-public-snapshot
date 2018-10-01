import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import React from 'react'
import { formatUserCount } from '../../productSubscription/helpers'

/**
 * Displays a text label with the product name (e.g., "Sourcegraph Enterprise") and user count for the
 * subscription.
 */
export const ProductSubscriptionLabel: React.SFC<{
    productSubscription: { plan: Pick<GQL.IProductPlan, 'fullProductName' | 'title'> | null } & Pick<
        GQL.IProductSubscription,
        'userCount'
    >

    planField?: 'fullProductName' | 'title'

    className?: string
}> = ({ productSubscription, planField = 'fullProductName', className = '' }) => (
    <span className={className}>
        {productSubscription.plan && productSubscription.userCount ? (
            <>
                {productSubscription.plan[planField]} ({formatUserCount(productSubscription.userCount)})
            </>
        ) : (
            <span className="text-muted font-italic">No plan selected</span>
        )}
    </span>
)

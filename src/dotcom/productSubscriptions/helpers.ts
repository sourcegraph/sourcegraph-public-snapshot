import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'

/**
 * Mirrors the GraphQL type ProductSubscriptionInput (but uses a plan object instead of ID).
 */
export interface ProductSubscriptionInput
    extends Pick<
            GQL.IProductSubscriptionInput,
            Exclude<keyof GQL.IProductSubscriptionInput, '__typename' | 'billingPlanID'>
        > {
    plan: GQL.IProductPlan
}

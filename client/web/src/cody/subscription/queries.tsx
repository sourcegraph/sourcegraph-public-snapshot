import { gql } from '@sourcegraph/http-client'

export const USER_CODY_PLAN = gql`
    query UserCodyPlan {
        currentUser {
            id
            codySubscription {
                status
                plan
                applyProRateLimits
                currentPeriodStartAt
                currentPeriodEndAt
                cancelAtPeriodEnd
            }
        }
    }
`

export const USER_CODY_USAGE = gql`
    query UserCodyUsage {
        currentUser {
            id
            codyCurrentPeriodChatUsage
            codyCurrentPeriodCodeUsage
            codyCurrentPeriodChatLimit
            codyCurrentPeriodCodeLimit
            codySubscription {
                status
                plan
                applyProRateLimits
                currentPeriodStartAt
                currentPeriodEndAt
                cancelAtPeriodEnd
            }
        }
    }
`

export const CHANGE_CODY_PLAN = gql`
    mutation ChangeCodyPlan($id: ID!, $pro: Boolean!) {
        changeCodyPlan(user: $id, pro: $pro) {
            id
            codySubscription {
                status
                plan
                applyProRateLimits
                currentPeriodStartAt
                currentPeriodEndAt
                cancelAtPeriodEnd
            }
        }
    }
`

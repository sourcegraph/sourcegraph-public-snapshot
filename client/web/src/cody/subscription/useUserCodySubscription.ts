import type { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import type { UserCodyPlanResult, UserCodyPlanVariables } from '../../graphql-operations'

export type UserCodySubscription = Omit<
    NonNullable<NonNullable<NonNullable<UserCodyPlanResult['currentUser']>>['codySubscription']>,
    '__typename'
>

const USER_CODY_PLAN = gql`
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

export const useUserCodySubscription = (): QueryResult<UserCodyPlanResult, UserCodyPlanVariables> =>
    useQuery<UserCodyPlanResult, UserCodyPlanVariables>(USER_CODY_PLAN, {})

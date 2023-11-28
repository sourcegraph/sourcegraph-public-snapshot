import { gql } from '@sourcegraph/http-client'

export const USER_CODY_PLAN = gql`
    query UserCodyPlan {
        currentUser {
            id
            codyProEnabled
            codyProEnabledAt
        }
    }
`

export const CHANGE_CODY_PLAN = gql`
    mutation ChangeCodyPlan($id: ID!, $pro: Boolean!) {
        changeCodyPlan(user: $id, pro: $pro) {
            id
            codyProEnabled
            codyProEnabledAt
        }
    }
`

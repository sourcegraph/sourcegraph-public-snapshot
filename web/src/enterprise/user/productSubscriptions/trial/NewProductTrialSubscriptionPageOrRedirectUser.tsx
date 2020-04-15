import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { UserSubscriptionsNewProductTrialSubscriptionPage } from './UserSubscriptionsNewProductTrialSubscriptionPage'
import { ThemeProps } from '../../../../../../shared/src/theme'

interface Props extends RouteComponentProps<{}>, ThemeProps {
    authenticatedUser: GQL.IUser | null
}

/**
 * Displays or redirects to the new product trial subscription page.
 *
 * For authenticated viewers, it redirects to the page under their user account.
 *
 * For unauthenticated viewers, it displays a page that lets them create a trial license without
 * signing in. This friendlier behavior for unauthed viewers (compared to dumping them on a sign-in
 * page) is the reason why this component exists.
 */
export const NewProductTrialSubscriptionPageOrRedirectUser: React.FunctionComponent<Props> = props =>
    props.authenticatedUser ? (
        <Redirect to={`${props.authenticatedUser.settingsURL!}/subscriptions/new-trial`} />
    ) : (
        <div className="container w-75 mt-4">
            <UserSubscriptionsNewProductTrialSubscriptionPage {...props} user={null} />
        </div>
    )

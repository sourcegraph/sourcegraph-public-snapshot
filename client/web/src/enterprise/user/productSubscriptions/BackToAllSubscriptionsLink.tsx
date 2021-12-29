import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import React from 'react'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { RouterLink } from '@sourcegraph/wildcard'

export const BackToAllSubscriptionsLink: React.FunctionComponent<{ user: Pick<GQL.IUser, 'settingsURL'> }> = ({
    user,
}) => (
    <RouterLink to={`${user.settingsURL!}/subscriptions`} className="btn btn-link btn-sm mb-3">
        <ArrowLeftIcon className="icon-inline" /> All subscriptions
    </RouterLink>
)

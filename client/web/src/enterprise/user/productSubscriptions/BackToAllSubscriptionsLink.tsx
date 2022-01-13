import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { Button } from '@sourcegraph/wildcard'

export const BackToAllSubscriptionsLink: React.FunctionComponent<{ user: Pick<GQL.IUser, 'settingsURL'> }> = ({
    user,
}) => (
    <Button to={`${user.settingsURL!}/subscriptions`} className="mb-3" variant="link" size="sm" as={Link}>
        <ArrowLeftIcon className="icon-inline" /> All subscriptions
    </Button>
)

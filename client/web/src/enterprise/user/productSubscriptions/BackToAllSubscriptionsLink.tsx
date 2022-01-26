import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'
import React from 'react'

import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, Link } from '@sourcegraph/wildcard'

export const BackToAllSubscriptionsLink: React.FunctionComponent<{ user: Pick<GQL.IUser, 'settingsURL'> }> = ({
    user,
}) => (
    <Button to={`${user.settingsURL!}/subscriptions`} className="mb-3" variant="link" size="sm" as={Link}>
        <ArrowLeftIcon className="icon-inline" /> All subscriptions
    </Button>
)

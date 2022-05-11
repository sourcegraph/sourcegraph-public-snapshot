import React from 'react'

import ArrowLeftIcon from 'mdi-react/ArrowLeftIcon'

import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, Link, Icon } from '@sourcegraph/wildcard'

export const BackToAllSubscriptionsLink: React.FunctionComponent<
    React.PropsWithChildren<{ user: Pick<GQL.IUser, 'settingsURL'> }>
> = ({ user }) => (
    <Button to={`${user.settingsURL!}/subscriptions`} className="mb-3" variant="link" size="sm" as={Link}>
        <Icon as={ArrowLeftIcon} /> All subscriptions
    </Button>
)

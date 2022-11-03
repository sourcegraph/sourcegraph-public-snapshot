import React from 'react'

import { mdiArrowLeft } from '@mdi/js'

import { Button, Link, Icon } from '@sourcegraph/wildcard'

import { UserAreaUserFields } from '../../../graphql-operations'

export const BackToAllSubscriptionsLink: React.FunctionComponent<
    React.PropsWithChildren<{ user: Pick<UserAreaUserFields, 'settingsURL'> }>
> = ({ user }) => (
    <Button to={`${user.settingsURL!}/subscriptions`} className="mb-3" variant="link" size="sm" as={Link}>
        <Icon aria-hidden={true} svgPath={mdiArrowLeft} /> All subscriptions
    </Button>
)

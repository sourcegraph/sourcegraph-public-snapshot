import React from 'react'

import { mdiArrowLeft } from '@mdi/js'

import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, Link, Icon } from '@sourcegraph/wildcard'

export const BackToAllSubscriptionsLink: React.FunctionComponent<
    React.PropsWithChildren<{ user: Pick<GQL.IUser, 'settingsURL'> }>
> = ({ user }) => (
    <Button to={`${user.settingsURL!}/subscriptions`} className="mb-3" variant="link" size="sm" as={Link}>
        <Icon aria-hidden={true} svgPath={mdiArrowLeft} /> All subscriptions
    </Button>
)

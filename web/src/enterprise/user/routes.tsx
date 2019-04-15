import React from 'react'
import { Redirect } from 'react-router'
import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute } from '../../user/area/UserArea'

export const enterpriseUserAreaRoutes: ReadonlyArray<UserAreaRoute> = [
    ...userAreaRoutes,

    // Redirect from previous /users/:username/subscriptions -> /users/:username/settings/subscriptions.
    {
        path: '/subscriptions/:page*',
        render: props => (
            <Redirect
                to={`${props.url}/settings/subscriptions${
                    props.match.params.page ? `/${props.match.params.page}` : ''
                }`}
            />
        ),
    },
]

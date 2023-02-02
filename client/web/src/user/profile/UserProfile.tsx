import React from 'react'

import { formatDistanceToNowStrict } from 'date-fns'

import { UserAreaRouteContext } from '../area/UserArea'

export const UserProfile: React.FunctionComponent<Pick<UserAreaRouteContext, 'user'>> = ({ user }) => {
    const userData: {
        name: string
        value: string
        visible: boolean
    }[] = [
        {
            name: 'Username',
            value: user.username,
            visible: true,
        },
        {
            name: 'Display name',
            value: user.displayName || 'Not set',
            visible: !!user.displayName,
        },
        {
            name: 'User since',
            value: formatDistanceToNowStrict(new Date(user.createdAt), { addSuffix: true }),
            visible: true,
        },
    ]

    return (
        <dl>
            {userData.map(data =>
                data.visible ? (
                    <>
                        <dt>{data.name}</dt>
                        <dd>{data.value}</dd>
                    </>
                ) : null
            )}
        </dl>
    )
}

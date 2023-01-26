import React from 'react'

import { formatDistanceToNowStrict } from 'date-fns'

import { UserAreaRouteContext } from '../area/UserArea'

export const UserProfile: React.FunctionComponent<
    Pick<UserAreaRouteContext, 'user'> & { isSourcegraphDotCom: boolean }
> = ({ user, isSourcegraphDotCom }) => {
    const primaryEmail = user.emails.find(email => email.isPrimary)?.email

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
        {
            name: 'Email',
            value: primaryEmail || 'Not set',
            visible: !!primaryEmail && !isSourcegraphDotCom, // Don't show email on Sourcegraph.com
        },
        {
            name: 'Site admin',
            value: user.siteAdmin ? 'Yes' : 'No',
            visible: user.siteAdmin && !isSourcegraphDotCom, // Don't show site admin status on Sourcegraph.com
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

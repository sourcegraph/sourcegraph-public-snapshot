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
        condition?: boolean
    }[] = [
        {
            name: 'Username',
            value: user.username,
        },
        {
            name: 'Display name',
            value: user.displayName || 'Not set',
            condition: !!user.displayName,
        },
        {
            name: 'User since',
            value: formatDistanceToNowStrict(new Date(user.createdAt), { addSuffix: true }),
        },
        {
            name: 'Email',
            value: primaryEmail || 'Not set',
            condition: !!primaryEmail && !isSourcegraphDotCom, // Don't show email on Sourcegraph.com
        },
        {
            name: 'Site admin',
            value: user.siteAdmin ? 'Yes' : 'No',
            condition: user.siteAdmin && !isSourcegraphDotCom, // Don't show site admin status on Sourcegraph.com
        },
    ]

    return (
        <dl>
            {userData.map(data =>
                data.condition !== false ? (
                    <>
                        <dt>{data.name}</dt>
                        <dd>{data.value}</dd>
                    </>
                ) : null
            )}
        </dl>
    )
}

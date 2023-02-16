import { Fragment, FC } from 'react'

import { formatDistanceToNowStrict } from 'date-fns'

import { UserAreaRouteContext } from '../area/UserArea'

export const UserProfile: FC<Pick<UserAreaRouteContext, 'user'>> = ({ user }) => {
    const primaryEmail = user.emails?.find(email => email.isPrimary)?.email

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
            visible: !!primaryEmail,
        },
    ]

    return (
        <dl>
            {userData.map(data =>
                data.visible ? (
                    <Fragment key={data.name}>
                        <dt>{data.name}</dt>
                        <dd>{data.value}</dd>
                    </Fragment>
                ) : null
            )}
        </dl>
    )
}

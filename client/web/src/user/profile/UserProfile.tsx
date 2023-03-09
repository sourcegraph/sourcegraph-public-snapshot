import { Fragment, FC, useState, useMemo } from 'react'

import { formatDistanceToNowStrict } from 'date-fns'

import { Button } from '@sourcegraph/wildcard'

import { UserAreaRouteContext } from '../area/UserArea'

type userData<T> = {
    name: string
    value: T
    visible: boolean
    isList: false
} | {
    name: string
    value: T[]
    visible: boolean
    isList: true
}

export const UserProfile: FC<Pick<UserAreaRouteContext, 'user'>> = ({ user }) => {
    // const [roles, setRoles] = useState(user.roles.nodes)
    const primaryEmail = user.emails?.find(email => email.isPrimary)?.email
    const roles = useMemo(() => {
        const roleNames = user.roles.nodes.map(role => role.name)
        return roleNames
    }, [user.roles])



    const userData: userData<string>[] = [
        {
            name: 'Username',
            value: user.username,
            visible: true,
            isList: false,
        },
        {
            name: 'Display name',
            value: user.displayName || 'Not set',
            visible: !!user.displayName,
            isList: false,
        },
        {
            name: 'User since',
            value: formatDistanceToNowStrict(new Date(user.createdAt), { addSuffix: true }),
            visible: true,
            isList: false,
        },
        {
            name: 'Email',
            value: primaryEmail || 'Not set',
            visible: !!primaryEmail,
            isList: false,
        },
        // probably skip pagination for starship: add a todo!
        {
            name: 'Roles',
            value: roles,
            visible: true,
            isList: true,
        }
    ]

    return (
        <dl>
            {userData.map(data =>
                data.visible ? (
                    <Fragment key={data.name}>
                        <dt>{data.name}</dt>
                        <dd>
                            {data.isList ? (
                                <ul>
                                    {data.value.map((value, index) => <li key={index}>{value}</li>)}
                                    {user.roles.pageInfo.hasNextPage && <Button variant='link'>Show More</Button>}
                                </ul>
                            ) : <>{data.value}</>}
                        </dd>
                    </Fragment>
                ) : null
            )}
        </dl>
    )
}

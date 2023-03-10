import { Fragment, FC } from 'react'

import { formatDistanceToNowStrict } from 'date-fns'

import { UserAreaRouteContext } from '../area/UserArea'

type UserData =
    | {
          name: string
          value: string
          visible: boolean
          isList: false
      }
    | {
          name: string
          value: string[]
          visible: boolean
          isList: true
      }

export const UserProfile: FC<Pick<UserAreaRouteContext, 'user'>> = ({ user }) => {
    const primaryEmail = user.emails?.find(email => email.isPrimary)?.email
    const roles = user.roles.nodes.map(role => role.name)

    const userData: UserData[] = [
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
        {
            name: 'Roles',
            value: roles,
            visible: true,
            isList: true,
        },
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
                                    {data.value.map((value, index) => (
                                        <li key={value}>{value}</li>
                                    ))}
                                </ul>
                            ) : (
                                <>{data.value}</>
                            )}
                        </dd>
                    </Fragment>
                ) : null
            )}
        </dl>
    )
}

import classNames from 'classnames'
import AccountCircleIcon from 'mdi-react/AccountCircleIcon'
import React from 'react'

import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

interface Props {
    size?: number
    user: {
        avatarURL: string | null
        displayName?: string
        username: string
    }
    className?: string
    ['data-tooltip']?: string
    targetID?: string
    alt?: string
}

/**
 * UserAvatar displays the avatar of a user.
 */
export const UserAvatar: React.FunctionComponent<Props> = ({
    size,
    user,
    className,
    targetID,
    // Exclude children since neither <img /> nor mdi-react icons receive them
    children,
    ...otherProps
}) => {
    const [isRedesignEnabled] = useRedesignToggle()

    if (user?.avatarURL) {
        let url = user.avatarURL
        try {
            const urlObject = new URL(user.avatarURL)
            if (size) {
                urlObject.searchParams.set('s', size.toString())
            }
            url = urlObject.href
        } catch {
            // noop
        }
        return (
            <img
                className={classNames(isRedesignEnabled ? 'user-avatar-refresh' : 'user-avatar', className)}
                src={url}
                id={targetID}
                alt=""
                role="presentation"
                {...otherProps}
            />
        )
    }

    if (isRedesignEnabled) {
        const name = user?.displayName || user.username
        const [firstName, lastName] = name.split(' ').map((name: string) => name.split('')[0])

        return (
            <div className={classNames('user-avatar-refresh__placeholder', className)}>
                {firstName}
                {lastName}
            </div>
        )
    }

    return <AccountCircleIcon className={classNames('user-avatar', className)} id={targetID} {...otherProps} />
}

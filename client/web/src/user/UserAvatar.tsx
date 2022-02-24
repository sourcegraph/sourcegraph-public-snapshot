import classNames from 'classnames'
import React from 'react'

import { Maybe } from '@sourcegraph/shared/src/graphql-operations'
import { Icon } from '@sourcegraph/wildcard'

import styles from './UserAvatar.module.scss'

interface Props {
    size?: number
    user: {
        avatarURL: Maybe<string>
        displayName: Maybe<string>
        username?: Maybe<string>
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
            <Icon as="span">
                <img
                    className={classNames(styles.userAvatar, className)}
                    src={url}
                    id={targetID}
                    alt=""
                    role="presentation"
                    {...otherProps}
                />
            </Icon>
        )
    }

    const name = user?.displayName || user?.username || ''
    const getInitials = (fullName: string): string => {
        const names = fullName.split(' ')
        const initials = names.map(name => name.charAt(0).toLowerCase())
        if (initials.length > 1) {
            return `${initials[0]}${initials[initials.length - 1]}`
        }
        return initials[0]
    }

    return (
        <Icon id={targetID} className={classNames(styles.userAvatar, className)} as="span">
            {getInitials(name)}
        </Icon>
    )
}

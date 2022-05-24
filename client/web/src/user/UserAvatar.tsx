import React from 'react'

import classNames from 'classnames'

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
    /**
     * Whether to render with icon-inline className
     */
    inline?: boolean
}

/**
 * UserAvatar displays the avatar of a user.
 */
export const UserAvatar: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    size,
    user,
    className,
    targetID,
    inline,
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

        const imgProps = {
            className: classNames(styles.userAvatar, className),
            src: url,
            id: targetID,
            role: 'presentation',
            ...otherProps,
        }

        if (inline) {
            return <Icon as="img" alt="" aria-label="User avatar" {...imgProps} />
        }

        return <img alt="" {...imgProps} />
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

    const props = {
        id: targetID,
        className: classNames(styles.userAvatar, className),
        children: <span className={styles.initials}>{getInitials(name)}</span>,
    }

    if (inline) {
        return <Icon role="img" as="div" aria-label="User avatar" {...props} />
    }

    return <div {...props} />
}

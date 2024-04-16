import React from 'react'

import classNames from 'classnames'

import { type ForwardReferenceComponent, Icon } from '@sourcegraph/wildcard'

import type { Maybe } from '../graphql-operations'

import styles from './UserAvatar.module.scss'

export interface UserAvatarData {
    avatarURL: Maybe<string>
    displayName: Maybe<string>
    username?: Maybe<string>
}

interface Props {
    size?: number
    user: UserAvatarData
    className?: string
    targetID?: string
    alt?: string
    /**
     * Whether to render with icon-inline className
     */
    inline?: boolean
    capitalizeInitials?: boolean
}

/**
 * UserAvatar displays the avatar of a user.
 */
export const UserAvatar = React.forwardRef(function UserAvatar(
    {
        size,
        user,
        className,
        targetID,
        inline,
        // Exclude children since neither <img /> nor mdi-react icons receive them
        children,
        capitalizeInitials,
        ...otherProps
    }: React.PropsWithChildren<Props>,
    reference
) {
    if (user?.avatarURL) {
        let url = user.avatarURL
        try {
            const urlObject = new URL(user.avatarURL)
            // Add a size param for non-data URLs. This will resize the image
            // if it is hosted on certain places like Gravatar and GitHub.
            if (size && !user.avatarURL.startsWith('data:')) {
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
            width: size,
            ...otherProps,
        }

        if (inline) {
            return (
                <Icon ref={reference as React.ForwardedRef<SVGSVGElement>} as="img" aria-hidden={true} {...imgProps} />
            )
        }

        return <img ref={reference} alt="" {...imgProps} />
    }

    const name = user?.displayName || user?.username || ''
    const getInitials = (fullName: string): string => {
        const names = fullName.split(' ')
        const initials = names.map(name => transformInitial(name.charAt(0), capitalizeInitials))
        if (initials.length > 1) {
            return `${initials[0]}${initials.at(-1)}`
        }
        return initials[0]
    }

    const sharedProps = {
        id: targetID,
        className: classNames(styles.userAvatar, className),
        children: <span className={styles.initials}>{getInitials(name)}</span>,
    }

    if (inline) {
        return (
            <Icon ref={reference as React.ForwardedRef<SVGSVGElement>} as="div" aria-hidden={true} {...sharedProps} />
        )
    }

    return <div ref={reference} {...sharedProps} />
}) as ForwardReferenceComponent<'img', React.PropsWithChildren<Props>>
UserAvatar.displayName = 'UserAvatar'

const transformInitial = (char: string, capitalize?: boolean): string =>
    capitalize ? char.toUpperCase() : char.toLowerCase()

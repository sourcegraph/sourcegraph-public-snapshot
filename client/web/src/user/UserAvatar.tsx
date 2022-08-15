import React from 'react'

import classNames from 'classnames'

import { Maybe } from '@sourcegraph/shared/src/graphql-operations'
import { ForwardReferenceComponent, Icon } from '@sourcegraph/wildcard'

import styles from './UserAvatar.module.scss'

interface Props {
    size?: number
    user: {
        avatarURL: Maybe<string>
        displayName: Maybe<string>
        username?: Maybe<string>
    }
    className?: string
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
// eslint-disable-next-line react/display-name
export const UserAvatar = React.forwardRef(
    (
        {
            size,
            user,
            className,
            targetID,
            inline,
            // Exclude children since neither <img /> nor mdi-react icons receive them
            children,
            ...otherProps
        }: React.PropsWithChildren<Props>,
        reference
    ) => {
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
                return (
                    <Icon
                        ref={reference as React.ForwardedRef<SVGSVGElement>}
                        as="img"
                        aria-hidden={true}
                        {...imgProps}
                    />
                )
            }

            return <img ref={reference} alt="" {...imgProps} />
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
            return <Icon ref={reference as React.ForwardedRef<SVGSVGElement>} as="div" aria-hidden={true} {...props} />
        }

        return <div ref={reference} {...props} />
    }
) as ForwardReferenceComponent<'img', React.PropsWithChildren<Props>>
UserAvatar.displayName = 'UserAvatar'

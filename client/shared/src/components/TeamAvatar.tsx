import * as React from 'react'

import classNames from 'classnames'

import { type ForwardReferenceComponent, Icon } from '@sourcegraph/wildcard'

import type { Maybe } from '../graphql-operations'

import styles from './TeamAvatar.module.scss'

export interface TeamAvatarProps {
    team: {
        name: Maybe<string>
        avatarURL: Maybe<string>
        displayName: Maybe<string>
    }

    size?: number

    className?: string
    targetID?: string
    alt?: string
    /**
     * Whether to render with icon-inline className
     */
    inline?: boolean
}

/**
 * TeamAvatar displays the avatar of a team.
 */
export const TeamAvatar = React.forwardRef(function TeamAvatar(
    {
        size,
        team,
        className,
        targetID,
        inline,
        // Exclude children since neither <img /> nor mdi-react icons receive them
        children,
        ...otherProps
    }: React.PropsWithChildren<TeamAvatarProps>,
    reference
) {
    if (team.avatarURL) {
        let url = team.avatarURL
        try {
            const urlObject = new URL(team.avatarURL)
            if (size) {
                urlObject.searchParams.set('s', size.toString())
            }
            url = urlObject.href
        } catch {
            // noop
        }

        const imgProps = {
            className: classNames(styles.teamAvatar, className),
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
                    alt=""
                    {...imgProps}
                />
            )
        }

        return <img ref={reference} alt="" {...imgProps} />
    }

    const name = team.displayName || team.name || ''
    const getInitials = (fullName: string): string => {
        const names = fullName.includes(' ')
            ? fullName.split(' ')
            : fullName.includes('-')
            ? fullName.split('-')
            : fullName.split('.')
        const initials = names.map(name => name.charAt(0).toLowerCase())
        if (initials.length > 1) {
            return `${initials[0]}${initials.at(-1)}`
        }
        return initials[0]
    }

    const sharedProps = {
        id: targetID,
        className: classNames(styles.teamAvatar, className),
        children: <span className={styles.initials}>{getInitials(name)}</span>,
    }

    if (inline) {
        return (
            <Icon ref={reference as React.ForwardedRef<SVGSVGElement>} as="span" aria-hidden={true} {...sharedProps} />
        )
    }

    return <div ref={reference} {...sharedProps} />
}) as ForwardReferenceComponent<'img', React.PropsWithChildren<TeamAvatarProps>>

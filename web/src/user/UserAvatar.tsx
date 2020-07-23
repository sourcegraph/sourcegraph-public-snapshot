import React from 'react'

interface Props {
    size?: number
    user: {
        avatarURL: string | null
    }
    className?: string
    ['data-tooltip']?: string
}

/**
 * UserAvatar displays the avatar of a user.
 */
export const UserAvatar: React.FunctionComponent<Props> = ({ size, user, className, ...otherProps }) => {
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
        return <img className={`user-avatar ${className || ''}`} src={url} {...otherProps} />
    }
    return null
}

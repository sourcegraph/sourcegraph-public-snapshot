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
export const UserAvatar: React.SFC<Props> = ({ size, user, className, ...otherProps }) => {
    if (user && user.avatarURL) {
        let url = user.avatarURL
        try {
            const urlObj = new URL(user.avatarURL)
            if (size) {
                urlObj.searchParams.set('s', size.toString())
            }
            url = urlObj.href
        } catch (e) {
            // noop
        }
        return <img className={`avatar-icon ${className || ''}`} src={url} {...otherProps} />
    }
    return null
}

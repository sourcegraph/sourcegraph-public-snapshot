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
            const urlObj = new URL(user.avatarURL)
            if (size) {
                urlObj.searchParams.set('s', size.toString())
            }
            url = urlObj.href
        } catch (e) {
            // noop
        }
        return <img className={`user-avatar ${className || ''}`} src={url} {...otherProps} />
    }
    return null
}

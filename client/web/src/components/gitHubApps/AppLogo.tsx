import React, { useState } from 'react'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'

/**
 * Displays the GitHub App or installation account logo, or a
 * fallback icon of the accounts initials when the image isn't available (this
 * happens when the image is locked behind a login we don't want to
 * authenticate here)
 */
interface AppLogoProps {
    /**
     * GitHub App logo or GitHub App installation account avatar
     */
    src: string | undefined
    /**
     * GitHub App name or GitHub App installation account login name
     */
    name: string
    /**
     * Specified alt string when used for non-GitHub App's specifically,
     * e.g. a GitHub App's installation account
     */
    alt?: string
    className: string
}

export const AppLogo: React.FC<AppLogoProps> = ({ src, name, alt = 'App logo', className }) => {
    const [fallbackImage, setFallbackImage] = useState<boolean>(false)

    return !fallbackImage ? (
        <img
            className={className}
            src={src}
            alt={alt}
            aria-hidden={true}
            // On certain code hosts, the image resource may be locked behind a login
            // screen. It's not practical to authenticate the user in this context, so instead,
            // we catch when there's an error loading the image and toggle the component
            // state to render a fallback icon instead.
            onError={() => setFallbackImage(true)}
        />
    ) : (
        <UserAvatar user={{ avatarURL: null, username: name, displayName: name }} className={className} />
    )
}

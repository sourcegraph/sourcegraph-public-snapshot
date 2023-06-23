import React, { useState } from 'react'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'

interface AppLogoProps {
    src: string | undefined
    name: string
    className: string
}

export const AppLogo: React.FC<AppLogoProps> = ({ src, name, className }) => {
    const [fallbackImage, setFallbackImage] = useState<boolean>(false)

    return !fallbackImage ? (
        <img
            className={className}
            src={src}
            alt="App logo"
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

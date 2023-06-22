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
            onError={() => {
                setFallbackImage(true)
            }}
        />
    ) : (
        <UserAvatar user={{ avatarURL: null, username: name, displayName: name }} className={className} />
    )
}

import React, { useState } from 'react'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'

import { BatchChangesCodeHostFields } from '../../graphql-operations'

import { GitHubApp } from './GitHubAppCard'

interface AppLogoProps {
    app: GitHubApp | BatchChangesCodeHostFields['commitSigningConfiguration']
    className: string
}

export const AppLogo: React.FC<AppLogoProps> = ({ app, className }) => {
    const [fallbackImage, setFallbackImage] = useState<boolean>(false)

    return !fallbackImage ? (
        <img
            className={className}
            src={app!.logo}
            alt="App logo"
            aria-hidden={true}
            onError={() => {
                setFallbackImage(true)
            }}
        />
    ) : (
        <UserAvatar user={{ avatarURL: null, username: app!.name, displayName: app!.name }} className={className} />
    )
}

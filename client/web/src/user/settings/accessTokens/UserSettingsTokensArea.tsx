import React, { useCallback, useState } from 'react'

import { Route, Routes } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { NotFoundPage } from '../../../components/HeroPage'
import type { CreateAccessTokenResult } from '../../../graphql-operations'
import type { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { UserSettingsCreateAccessTokenCallbackPage } from './UserSettingsCreateAccessTokenCallbackPage'
import { UserSettingsCreateAccessTokenPage } from './UserSettingsCreateAccessTokenPage'
import { UserSettingsTokensPage } from './UserSettingsTokensPage'

interface Props
    extends Pick<UserSettingsAreaRouteContext, 'user' | 'authenticatedUser'>,
        TelemetryProps,
        TelemetryV2Props {
    isSourcegraphDotCom: boolean
}

export const UserSettingsTokensArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [newToken, setNewToken] = useState<CreateAccessTokenResult['createAccessToken'] | undefined>()

    const onDidPresentNewToken = useCallback(() => {
        setNewToken(undefined)
    }, [])

    return (
        <Routes>
            <Route
                path="new"
                element={<UserSettingsCreateAccessTokenPage {...props} onDidCreateAccessToken={setNewToken} />}
            />
            <Route
                path="new/callback"
                element={<UserSettingsCreateAccessTokenCallbackPage {...props} onDidCreateAccessToken={setNewToken} />}
            />
            <Route
                path=""
                element={
                    <UserSettingsTokensPage
                        {...props}
                        newToken={newToken}
                        onDidPresentNewToken={onDidPresentNewToken}
                    />
                }
            />
            <Route path="*" element={<NotFoundPage pageType="settings" />} />
        </Routes>
    )
}

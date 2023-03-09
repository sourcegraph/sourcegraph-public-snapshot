import React, { useCallback, useState } from 'react'

import { Route, Routes } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { NotFoundPage } from '../../../components/HeroPage'
import { CreateAccessTokenResult } from '../../../graphql-operations'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { UserSettingsCreateAccessTokenCallbackPage } from './UserSettingsCreateAccessTokenCallbackPage'
import { UserSettingsCreateAccessTokenPage } from './UserSettingsCreateAccessTokenPage'
import { UserSettingsTokensPage } from './UserSettingsTokensPage'

interface Props extends Pick<UserSettingsAreaRouteContext, 'user' | 'authenticatedUser'>, TelemetryProps {
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
            <Route element={<NotFoundPage pageType="settings" />} />
        </Routes>
    )
}

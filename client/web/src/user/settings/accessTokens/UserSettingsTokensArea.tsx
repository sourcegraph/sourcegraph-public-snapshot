import React, { useCallback, useState } from 'react'

import { Navigate, Route, Routes } from 'react-router-dom'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { NotFoundPage } from '../../../components/HeroPage'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import type { CreateAccessTokenResult } from '../../../graphql-operations'
import { PageRoutes } from '../../../routes.constants'
import type { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { UserSettingsCreateAccessTokenCallbackPage } from './UserSettingsCreateAccessTokenCallbackPage'
import { UserSettingsCreateAccessTokenPage } from './UserSettingsCreateAccessTokenPage'
import { UserSettingsTokensPage } from './UserSettingsTokensPage'

interface Props extends Pick<UserSettingsAreaRouteContext, 'user' | 'authenticatedUser'>, TelemetryProps {
    isSourcegraphDotCom: boolean
}

export const UserSettingsTokensArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [newToken, setNewToken] = useState<CreateAccessTokenResult['createAccessToken'] | undefined>()
    // Start with `true` to avoid redirecting before having a chance to check the real value of the flag.
    const [isCodyProEnabled] = useFeatureFlag('cody-pro', true)

    const onDidPresentNewToken = useCallback(() => {
        setNewToken(undefined)
    }, [])

    if (
        props.isSourcegraphDotCom &&
        !isCodyProEnabled &&
        props.authenticatedUser &&
        !props.authenticatedUser.completedPostSignup
    ) {
        const returnTo = window.location.href
        const params = new URLSearchParams()
        params.set('returnTo', returnTo)
        const navigateTo = PageRoutes.PostSignUp + '?' + params.toString()
        return <Navigate to={navigateTo.toString()} replace={true} />
    }
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

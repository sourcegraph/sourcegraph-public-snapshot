import React, { useCallback, useState } from 'react'

import { Route, RouteComponentProps, Switch } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { NotFoundPage } from '../../../components/HeroPage'
import { CreateAccessTokenResult } from '../../../graphql-operations'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { UserSettingsCreateAccessTokenCallbackPage } from './UserSettingsCreateAccessTokenCallbackPage'
import { UserSettingsCreateAccessTokenPage } from './UserSettingsCreateAccessTokenPage'
import { UserSettingsTokensPage } from './UserSettingsTokensPage'

interface Props
    extends Pick<UserSettingsAreaRouteContext, 'user' | 'authenticatedUser'>,
        Pick<RouteComponentProps<{}>, 'history' | 'location' | 'match'>,
        TelemetryProps {}

export const UserSettingsTokensArea: React.FunctionComponent<React.PropsWithChildren<Props>> = outerProps => {
    const [newToken, setNewToken] = useState<CreateAccessTokenResult['createAccessToken'] | undefined>()
    const onDidPresentNewToken = useCallback(() => {
        setNewToken(undefined)
    }, [])
    return (
        <Switch>
            <Route
                exact={true}
                path={outerProps.match.url + '/new'}
                render={props => (
                    <UserSettingsCreateAccessTokenPage
                        {...outerProps}
                        {...props}
                        onDidCreateAccessToken={setNewToken}
                    />
                )}
            />
            <Route
                exact={true}
                path={outerProps.match.url + '/new/callback'}
                render={props => (
                    <UserSettingsCreateAccessTokenCallbackPage
                        {...outerProps}
                        {...props}
                        onDidCreateAccessToken={setNewToken}
                    />
                )}
            />
            <Route
                exact={true}
                path={outerProps.match.url}
                render={props => (
                    <UserSettingsTokensPage
                        {...outerProps}
                        {...props}
                        newToken={newToken}
                        onDidPresentNewToken={onDidPresentNewToken}
                    />
                )}
            />
            <Route render={() => <NotFoundPage pageType="settings" />} key="hardcoded-key" />
        </Switch>
    )
}

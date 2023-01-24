import React, { useCallback, useState } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { RouteComponentProps, Switch } from 'react-router'
import { CompatRoute } from 'react-router-dom-v5-compat'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { HeroPage } from '../../../components/HeroPage'
import { CreateAccessTokenResult } from '../../../graphql-operations'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { UserSettingsCreateAccessTokenCallbackPage } from './UserSettingsCreateAccessTokenCallbackPage'
import { UserSettingsCreateAccessTokenPage } from './UserSettingsCreateAccessTokenPage'
import { UserSettingsTokensPage } from './UserSettingsTokensPage'

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)

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
            <CompatRoute
                exact={true}
                path={outerProps.match.url + '/new'}
                render={(props: RouteComponentProps) => (
                    <UserSettingsCreateAccessTokenPage
                        {...outerProps}
                        {...props}
                        onDidCreateAccessToken={setNewToken}
                    />
                )}
            />
            <CompatRoute
                exact={true}
                path={outerProps.match.url + '/new/callback'}
                render={(props: RouteComponentProps) => (
                    <UserSettingsCreateAccessTokenCallbackPage
                        {...outerProps}
                        {...props}
                        onDidCreateAccessToken={setNewToken}
                    />
                )}
            />
            <CompatRoute
                exact={true}
                path={outerProps.match.url}
                render={(props: RouteComponentProps) => (
                    <UserSettingsTokensPage
                        {...outerProps}
                        {...props}
                        newToken={newToken}
                        onDidPresentNewToken={onDidPresentNewToken}
                    />
                )}
            />
            <CompatRoute component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    )
}

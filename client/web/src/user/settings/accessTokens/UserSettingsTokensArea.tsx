import React, { useCallback, useState } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { HeroPage } from '../../../components/HeroPage'
import { CreateAccessTokenResult } from '../../../graphql-operations'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

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
            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    )
}

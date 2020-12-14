import React, { useCallback, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'
import { UserSettingsCreateAccessTokenPage } from './UserSettingsCreateAccessTokenPage'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { HeroPage } from '../../../components/HeroPage'
import { UserSettingsTokensPage } from './UserSettingsTokensPage'
import { CreateAccessTokenResult } from '../../../graphql-operations'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

interface Props
    extends Pick<UserSettingsAreaRouteContext, 'user' | 'authenticatedUser'>,
        Pick<RouteComponentProps<{}>, 'history' | 'location' | 'match'>,
        TelemetryProps {}

export const UserSettingsTokensArea: React.FunctionComponent<Props> = outerProps => {
    const [newToken, setNewToken] = useState<CreateAccessTokenResult['createAccessToken'] | undefined>()
    const onDidPresentNewToken = useCallback(() => {
        setNewToken(undefined)
    }, [])
    return (
        /* eslint-disable react/jsx-no-bind */
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
        /* eslint-enable react/jsx-no-bind */
    )
}

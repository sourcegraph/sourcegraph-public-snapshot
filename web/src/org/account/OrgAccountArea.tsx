import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { ErrorBoundary } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import { ThemeProps } from '../../theme'
import { OrgAreaPageProps } from '../area/OrgArea'
import { OrgAccountProfilePage } from './OrgAccountProfilePage'

const NotFoundPage = () => (
    <HeroPage
        icon={MapSearchIcon}
        title="404: Not Found"
        subtitle="Sorry, the requested organization page was not found."
    />
)

interface Props extends OrgAreaPageProps, RouteComponentProps<{}>, ThemeProps {
    location: H.Location
}

/**
 * Renders a layout of a sidebar and a content area to display pages related to
 * an organization's settings.
 */
export const OrgAccountArea: React.FunctionComponent<Props> = props => {
    if (props.location.pathname === props.match.path) {
        return <Redirect to={`${props.match.path}/profile`} />
    }
    if (!props.authenticatedUser) {
        return null
    }

    const transferProps: OrgAreaPageProps = {
        authenticatedUser: props.authenticatedUser,
        org: props.org,
        onOrganizationUpdate: props.onOrganizationUpdate,
        platformContext: props.platformContext,
        settingsCascade: props.settingsCascade,
    }
    return (
        <div className="mt-3">
            <ErrorBoundary location={props.location}>
                <React.Suspense fallback={<LoadingSpinner className="icon-inline m-2" />}>
                    <Switch>
                        <Route
                            path={`${props.match.path}/profile`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            exact={true}
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <OrgAccountProfilePage {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route component={NotFoundPage} />
                    </Switch>
                </React.Suspense>
            </ErrorBoundary>
        </div>
    )
}

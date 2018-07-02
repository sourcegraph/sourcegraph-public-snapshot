import CityIcon from '@sourcegraph/icons/lib/City'
import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import UserIcon from '@sourcegraph/icons/lib/User'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { NavLink } from 'react-router-dom'
import { HeroPage } from '../components/HeroPage'
import { eventLogger } from '../tracking/eventLogger'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionUsageOrganizationsPage } from './RegistryExtensionUsageOrganizationsPage'
import { RegistryExtensionUsageUsersPage } from './RegistryExtensionUsageUsersPage'

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {}

/** An area displaying information about usage and configuration of a registry extension. */
export class RegistryExtensionUsageArea extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionUsage')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-usage-area">
                <div className="btn-group mb-3">
                    <NavLink
                        className="btn btn-secondary"
                        activeClassName="active font-weight-bold"
                        to={`${this.props.match.path}/users`}
                        exact={true}
                        data-tooltip="Users with this extension enabled"
                    >
                        <UserIcon className="icon-inline" /> Users
                    </NavLink>
                    <NavLink
                        className="btn btn-secondary"
                        activeClassName="active font-weight-bold"
                        to={`${this.props.match.path}/organizations`}
                        exact={true}
                        data-tooltip="Organizations that enable this extension for members"
                    >
                        <CityIcon className="icon-inline" /> Organizations
                    </NavLink>
                </div>
                <Switch>
                    <Route
                        path={`${this.props.match.path}/users`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <RegistryExtensionUsageUsersPage {...routeComponentProps} {...this.props} />
                        )}
                    />
                    <Route
                        path={`${this.props.match.path}/organizations`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <RegistryExtensionUsageOrganizationsPage {...routeComponentProps} {...this.props} />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}

import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import * as React from 'react'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { Link, NavLink } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { eventLogger } from '../tracking/eventLogger'
import { ConfiguredExtensionsListProps, ConfiguredExtensionsPage } from './ConfiguredExtensionsPage'
import { RegistryPublisher } from './extension'
import { ExtensionsListViewMode, RegistryExtensionsList } from './RegistryExtensionsPage'

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

interface Props extends RouteComponentProps<{}>, ConfiguredExtensionsListProps {
    authenticatedUser: GQL.IUser | null

    /** The extension publisher. */
    publisher: {
        registryExtensions?: Pick<GQL.IRegistryExtensionConnection, 'url'>
    } & RegistryPublisher
}

/**
 * An area displaying extensions published by and used by a user or organization (which are both registry extension
 * publishers and extension configuration subjects).
 */
export class PublisherSubjectExtensionsArea extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('PublisherSubjectExtensionsArea')
    }

    public render(): JSX.Element | null {
        const publishedPath = `${this.props.match.path}/published`
        if (this.props.location.pathname === this.props.match.path) {
            return <Redirect to={publishedPath} />
        }

        const noun =
            this.props.publisher.__typename === 'User' ? this.props.publisher.username : this.props.publisher.name

        return (
            <div className="publisher-subject-extensions-area">
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <div className="btn-group mr-2">
                        <NavLink
                            className="btn btn-outline-primary"
                            activeClassName="active font-weight-bold"
                            to={publishedPath}
                            exact={true}
                            data-tooltip={`Extensions published by ${noun}`}
                        >
                            Published by {noun}
                        </NavLink>
                        <NavLink
                            className="btn btn-outline-primary"
                            activeClassName="active font-weight-bold"
                            to={`${this.props.match.path}/used`}
                            exact={true}
                            data-tooltip={`Extensions used by ${noun}`}
                        >
                            Used by {noun}
                        </NavLink>
                    </div>
                    <div>
                        <Link to="/registry" className="btn btn-outline-link">
                            Extension registry
                        </Link>{' '}
                        {this.props.subject &&
                            this.props.subject.settingsURL &&
                            this.props.subject.viewerCanAdminister && (
                                <Link className="btn btn-primary" to={this.props.subject.settingsURL}>
                                    <GearIcon className="icon-inline" /> Configure extensions
                                </Link>
                            )}
                    </div>
                </div>
                <Switch>
                    <Route
                        path={publishedPath}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <RegistryExtensionsList
                                {...routeComponentProps}
                                {...this.props}
                                publisher={this.props.publisher}
                                mode={ExtensionsListViewMode.List}
                            />
                        )}
                    />
                    <Route
                        path={`${this.props.match.path}/used`}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => (
                            <ConfiguredExtensionsPage
                                {...routeComponentProps}
                                {...this.props}
                                subject={this.props.subject}
                            />
                        )}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }
}

import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionsEmptyState } from './ExtensionsEmptyState'
import { ExtensionsList } from './ExtensionsList'

interface Props extends ExtensionsAreaRouteContext, RouteComponentProps<{}> {}

/** A page that displays overview information about the available extensions. */
export class ExtensionsOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('ExtensionsOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="container px-2 px-xl-0">
                <PageTitle title="Extensions" />
                <div className="py-3">
                    {!this.props.authenticatedUser && (
                        <div className="alert alert-info">
                            <Link to="/sign-in" className="btn btn-primary mr-2">
                                Sign in to add and configure extensions
                            </Link>
                            <small>An account is required.</small>
                        </div>
                    )}
                    <ExtensionsList
                        {...this.props}
                        emptyElement={<ExtensionsEmptyState />}
                        subject={this.props.subject}
                        settingsCascade={this.props.settingsCascade}
                    />
                </div>
            </div>
        )
    }
}

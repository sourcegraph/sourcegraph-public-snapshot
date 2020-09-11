import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionsList } from './ExtensionsList'

interface Props
    extends Pick<ExtensionsAreaRouteContext, 'authenticatedUser' | 'subject'>,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        SettingsCascadeProps {
    location: H.Location
    history: H.History
}

/** A page that displays overview information about the available extensions. */
export class ExtensionsOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('ExtensionsOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="container">
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
                        subject={this.props.subject}
                        settingsCascade={this.props.settingsCascade}
                    />
                </div>
            </div>
        )
    }
}

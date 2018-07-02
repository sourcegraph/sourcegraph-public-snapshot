import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionUsersList } from './RegistryExtensionUsersList'

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {}

/** A page that displays the list of users of an extension. */
export class RegistryExtensionUsageUsersPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionUsageUsers')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-usage-users-page">
                <PageTitle title={`Users of ${this.props.extension.extensionID}`} />
                <RegistryExtensionUsersList {...this.props} />
            </div>
        )
    }
}

import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionConfigurationSubjectsList } from './RegistryExtensionConfigurationSubjectsList'

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {}

/** A page that displays the list of organizations that enable an extension. */
export class RegistryExtensionUsageOrganizationsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionUsageOrganizations')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-usage-organizations-page">
                <PageTitle title={`Organizations using ${this.props.extension.extensionID}`} />
                <RegistryExtensionConfigurationSubjectsList {...this.props} />
            </div>
        )
    }
}

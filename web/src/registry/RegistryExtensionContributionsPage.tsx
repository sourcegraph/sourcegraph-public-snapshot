import InfoIcon from '@sourcegraph/icons/lib/Info'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionContributions } from './RegistryExtensionContributions'

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {}

/** A page that displays an extension's contributions. */
export class RegistryExtensionContributionsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionContributions')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-contributions-page">
                <PageTitle title={`Contributions from ${this.props.extension.extensionID}`} />
                <div className="d-flex align-items-center">
                    <h3 className="mb-0 mr-1">Contributions</h3>
                    <InfoIcon
                        className="icon-inline text-muted"
                        data-tooltip="The features provided by the extension (queried live)"
                    />
                </div>
                <RegistryExtensionContributions extension={this.props.extension} />
            </div>
        )
    }
}

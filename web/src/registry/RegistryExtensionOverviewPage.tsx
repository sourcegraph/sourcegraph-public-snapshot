import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../components/LinkOrSpan'
import { PageTitle } from '../components/PageTitle'
import { Timestamp } from '../components/time/Timestamp'
import { eventLogger } from '../tracking/eventLogger'
import { extensionIDPrefix } from './extension'
import { RegistryExtensionAreaPageProps } from './RegistryExtensionArea'
import { RegistryExtensionDescription } from './RegistryExtensionDescription'
import { RegistryExtensionViewerConfigurationStatus } from './RegistryExtensionViewerConfigurationStatus'

interface Props extends RegistryExtensionAreaPageProps, RouteComponentProps<{}> {}

/** A page that displays overview information about a registry extension. */
export class RegistryExtensionOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-overview-page row">
                <PageTitle title={this.props.extension.extensionID} />
                <div className="col-md-8">
                    <RegistryExtensionDescription extension={this.props.extension} />
                </div>
                <div className="col-md-4">
                    <RegistryExtensionViewerConfigurationStatus
                        {...this.props}
                        nonEmptyClassName="mb-3"
                        onUpdate={this.props.onDidUpdateExtension}
                    />
                    <small className="text-muted">
                        <dl className="mt-3">
                            {this.props.extension.publisher && (
                                <>
                                    <dt>Publisher</dt>
                                    <dd>
                                        {this.props.extension.publisher ? (
                                            <Link to={this.props.extension.publisher.url}>
                                                {extensionIDPrefix(this.props.extension.publisher)}
                                            </Link>
                                        ) : (
                                            'Unavailable'
                                        )}
                                    </dd>
                                </>
                            )}
                            {this.props.extension.registryName && (
                                <>
                                    <dt className={this.props.extension.publisher ? 'border-top pt-1' : ''}>
                                        Published on
                                    </dt>
                                    <dd>
                                        <LinkOrSpan to={this.props.extension.remoteURL}>
                                            {this.props.extension.registryName}
                                        </LinkOrSpan>
                                    </dd>
                                </>
                            )}
                            <dt className="border-top pt-1">Extension ID</dt>
                            <dd>{this.props.extension.extensionID}</dd>
                            {this.props.extension.updatedAt && (
                                <>
                                    <dt className="border-top pt-1">Last updated</dt>
                                    <dd>
                                        <Timestamp date={this.props.extension.updatedAt} />
                                    </dd>
                                </>
                            )}
                        </dl>
                    </small>
                </div>
            </div>
        )
    }
}

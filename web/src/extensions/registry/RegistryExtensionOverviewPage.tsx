import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../components/LinkOrSpan'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { eventLogger } from '../../tracking/eventLogger'
import { isErrorLike } from '../../util/errors'
import { extensionIDPrefix } from '../extension/extension'
import { ExtensionAreaPageProps } from '../extension/ExtensionArea'
import { ExtensionREADME } from '../extension/RegistryExtensionREADME'

interface Props extends ExtensionAreaPageProps, RouteComponentProps<{}> {}

/** A page that displays overview information about a registry extension. */
export class RegistryExtensionOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionOverview')
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-overview-page row">
                <PageTitle title={this.props.extension.id} />
                <div className="col-md-8">
                    <ExtensionREADME extension={this.props.extension} />
                </div>
                <div className="col-md-4">
                    <small className="text-muted">
                        <dl className="border-top pt-2">
                            {this.props.extension.registryExtension &&
                                this.props.extension.registryExtension.publisher && (
                                    <>
                                        <dt>Publisher</dt>
                                        <dd>
                                            {this.props.extension.registryExtension.publisher ? (
                                                <Link to={this.props.extension.registryExtension.publisher.url}>
                                                    {extensionIDPrefix(
                                                        this.props.extension.registryExtension.publisher
                                                    )}
                                                </Link>
                                            ) : (
                                                'Unavailable'
                                            )}
                                        </dd>
                                    </>
                                )}
                            {this.props.extension.registryExtension &&
                                this.props.extension.registryExtension.registryName && (
                                    <>
                                        <dt
                                            className={
                                                this.props.extension.registryExtension.publisher
                                                    ? 'border-top pt-2'
                                                    : ''
                                            }
                                        >
                                            Published on
                                        </dt>
                                        <dd>
                                            <LinkOrSpan
                                                to={this.props.extension.registryExtension.remoteURL}
                                                target={
                                                    this.props.extension.registryExtension.isLocal ? undefined : '_self'
                                                }
                                            >
                                                {this.props.extension.registryExtension.registryName}
                                            </LinkOrSpan>
                                        </dd>
                                    </>
                                )}
                            <dt className="border-top pt-2">Extension ID</dt>
                            <dd>{this.props.extension.id}</dd>
                            {this.props.extension.registryExtension &&
                                this.props.extension.registryExtension.updatedAt && (
                                    <>
                                        <dt className="border-top pt-2">Last updated</dt>
                                        <dd>
                                            <Timestamp date={this.props.extension.registryExtension.updatedAt} />
                                        </dd>
                                    </>
                                )}
                            <dt className="border-top pt-2">Resources</dt>
                            <dd className="border-bottom pb-2">
                                <Link
                                    to={`${this.props.extension.registryExtension!.url}/-/manifest`}
                                    className="d-block"
                                >
                                    Manifest (package.json)
                                </Link>
                                {this.props.extension.manifest &&
                                    !isErrorLike(this.props.extension.manifest) &&
                                    this.props.extension.manifest.url && (
                                        <a
                                            href={this.props.extension.manifest.url}
                                            rel="nofollow"
                                            target="_blank"
                                            className="d-block"
                                        >
                                            Source code (JavaScript)
                                        </a>
                                    )}
                            </dd>
                        </dl>
                    </small>
                </div>
            </div>
        )
    }
}

import { isObject } from 'lodash'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../../../packages/webapp/src/components/LinkOrSpan'
import { PageTitle } from '../../../../packages/webapp/src/components/PageTitle'
import { Timestamp } from '../../../../packages/webapp/src/components/time/Timestamp'
import { extensionIDPrefix } from '../../../../packages/webapp/src/extensions/extension/extension'
import { ExtensionAreaRouteContext } from '../../../../packages/webapp/src/extensions/extension/ExtensionArea'
import { ExtensionREADME } from '../../../../packages/webapp/src/extensions/extension/RegistryExtensionREADME'
import { eventLogger } from '../../../../packages/webapp/src/tracking/eventLogger'
import { isErrorLike } from '../../../../packages/webapp/src/util/errors'

interface Props extends ExtensionAreaRouteContext, RouteComponentProps<{}> {}

/** A page that displays overview information about a registry extension. */
export class RegistryExtensionOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionOverview')
    }

    public render(): JSX.Element | null {
        let repositoryURL: URL | undefined
        try {
            if (
                this.props.extension.manifest &&
                !isErrorLike(this.props.extension.manifest) &&
                this.props.extension.manifest.repository &&
                isObject(this.props.extension.manifest.repository) &&
                typeof this.props.extension.manifest.repository.url === 'string'
            ) {
                repositoryURL = new URL(this.props.extension.manifest.repository.url)
            }
        } catch (e) {
            // noop
        }

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
                                {repositoryURL && (
                                    <div className="d-flex">
                                        {repositoryURL.hostname === 'github.com' && (
                                            <GithubCircleIcon className="icon-inline" />
                                        )}
                                        <a
                                            href={repositoryURL.href}
                                            rel="nofollow noreferrer noopener"
                                            target="_blank"
                                            className="d-block"
                                        >
                                            Repository
                                        </a>
                                    </div>
                                )}
                            </dd>
                        </dl>
                    </small>
                </div>
            </div>
        )
    }
}

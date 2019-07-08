import { parseISO } from 'date-fns'
import maxDate from 'date-fns/max'
import { isObject, truncate } from 'lodash'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { ExtensionCategory } from '../../../../shared/src/schema/extensionSchema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { isDefined } from '../../../../shared/src/util/types'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { EventLogger } from '../../tracking/eventLogger'
import { extensionIDPrefix, extensionsQuery, urlToExtensionsQuery, validCategories } from './extension'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionREADME } from './RegistryExtensionREADME'

interface Props extends Pick<ExtensionAreaRouteContext, 'extension'> {
    eventLogger: Pick<EventLogger, 'logViewEvent'>
}

/** A page that displays overview information about a registry extension. */
export class RegistryExtensionOverviewPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        this.props.eventLogger.logViewEvent('RegistryExtension')
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

        let categories: ExtensionCategory[] | undefined
        if (
            this.props.extension.manifest &&
            !isErrorLike(this.props.extension.manifest) &&
            this.props.extension.manifest.categories
        ) {
            const cs = validCategories(this.props.extension.manifest.categories)
            if (cs && cs.length > 0) {
                categories = cs
            }
        }

        return (
            <div className="registry-extension-overview-page d-flex flex-wrap">
                <PageTitle title={this.props.extension.id} />
                <div className="registry-extension-overview-page__readme mr-3">
                    <ExtensionREADME extension={this.props.extension} />
                </div>
                <aside className="registry-extension-overview-page__sidebar">
                    {categories && (
                        <div className="mb-3">
                            <h3 className="mb-0">Categories</h3>
                            <ul className="list-inline registry-extension-overview-page__categories">
                                {categories.map((c, i) => (
                                    <li key={i} className="list-inline-item mb-2 small">
                                        <Link
                                            to={urlToExtensionsQuery(extensionsQuery({ category: c }))}
                                            className="rounded border p-1"
                                        >
                                            {c}
                                        </Link>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}
                    {this.props.extension.manifest &&
                        !isErrorLike(this.props.extension.manifest) &&
                        this.props.extension.manifest.tags &&
                        this.props.extension.manifest.tags.length > 0 && (
                            <div className="mb-3">
                                <h3 className="mb-0">Tags</h3>
                                <ul className="list-inline registry-extension-overview-page__tags">
                                    {this.props.extension.manifest.tags.map((t, i) => (
                                        <li key={i} className="list-inline-item mb-2 small">
                                            <Link
                                                to={urlToExtensionsQuery(extensionsQuery({ tag: t }))}
                                                className="rounded border p-1"
                                            >
                                                {truncate(t, { length: 24 })}
                                            </Link>
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        )}
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
                                (this.props.extension.registryExtension.updatedAt ||
                                    this.props.extension.registryExtension.publishedAt) && (
                                    <>
                                        <dt className="border-top pt-2">Last updated</dt>
                                        <dd>
                                            <Timestamp
                                                date={maxDate(
                                                    [
                                                        this.props.extension.registryExtension.updatedAt,
                                                        this.props.extension.registryExtension.publishedAt,
                                                    ]
                                                        .filter(isDefined)
                                                        .map(date => parseISO(date))
                                                )}
                                            />
                                        </dd>
                                    </>
                                )}
                            <dt className="border-top pt-2">Resources</dt>
                            <dd className="border-bottom pb-2">
                                {this.props.extension.registryExtension && (
                                    <Link
                                        to={`${this.props.extension.registryExtension.url}/-/manifest`}
                                        className="d-block"
                                    >
                                        Manifest (package.json)
                                    </Link>
                                )}
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
                                            <GithubCircleIcon className="icon-inline mr-1" />
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
                </aside>
            </div>
        )
    }
}

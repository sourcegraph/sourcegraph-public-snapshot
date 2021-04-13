import { parseISO } from 'date-fns'
import maxDate from 'date-fns/max'
import * as H from 'history'
import { isObject } from 'lodash'
import GithubIcon from 'mdi-react/GithubIcon'
import React, { useMemo, useEffect } from 'react'
import { Link } from 'react-router-dom'

import { ExtensionCategory, ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isEncodedImage } from '@sourcegraph/shared/src/util/icon'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { DefaultIcon } from '../icons'

import { extensionIDPrefix, extensionsQuery, urlToExtensionsQuery, validCategories } from './extension'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionReadme } from './RegistryExtensionReadme'

interface Props extends Pick<ExtensionAreaRouteContext, 'extension' | 'telemetryService' | 'isLightTheme'> {
    history: H.History
}

const RegistryExtensionOverviewIcon: React.FunctionComponent<Pick<Props, 'extension' | 'isLightTheme'>> = ({
    extension,
    isLightTheme,
}) => {
    const manifest: ExtensionManifest | undefined =
        extension.manifest && !isErrorLike(extension.manifest) ? extension.manifest : undefined

    const iconURL = useMemo(() => {
        let iconURL: URL | undefined

        try {
            if (isLightTheme) {
                if (manifest?.icon && isEncodedImage(manifest.icon)) {
                    iconURL = new URL(manifest.icon)
                }
            } else if (manifest?.iconDark && isEncodedImage(manifest.iconDark)) {
                iconURL = new URL(manifest.iconDark)
            } else if (manifest?.icon && isEncodedImage(manifest.icon)) {
                // fallback: show default icon on dark theme if dark icon isn't specified
                iconURL = new URL(manifest.icon)
            }
        } catch {
            // noop
        }

        return iconURL
    }, [manifest?.icon, manifest?.iconDark, isLightTheme])

    if (iconURL) {
        return <img className="registry-extension-overview-page__icon mb-3" src={iconURL.href} alt="" />
    }

    if (manifest?.publisher === 'sourcegraph') {
        return <DefaultIcon className="registry-extension-overview-page__icon mb-3" />
    }

    return null
}

/** A page that displays overview information about a registry extension. */
export const RegistryExtensionOverviewPage: React.FunctionComponent<Props> = ({
    telemetryService,
    extension,
    history,
    isLightTheme,
}) => {
    let repositoryURL: URL | undefined
    let categories: ExtensionCategory[] | undefined

    useEffect(() => {
        telemetryService.logViewEvent('AddExternalService')
    }, [telemetryService])

    try {
        if (
            extension.manifest &&
            !isErrorLike(extension.manifest) &&
            extension.manifest.repository &&
            isObject(extension.manifest.repository) &&
            typeof extension.manifest.repository.url === 'string'
        ) {
            repositoryURL = new URL(extension.manifest.repository.url)
        }
    } catch {
        // noop
    }

    if (extension.manifest && !isErrorLike(extension.manifest) && extension.manifest.categories) {
        const validatedCategories = validCategories(extension.manifest.categories)
        if (validatedCategories && validatedCategories.length > 0) {
            categories = validatedCategories
        }
    }

    return (
        <div className="registry-extension-overview-page d-flex flex-wrap">
            <PageTitle title={extension.id} />
            <div className="registry-extension-overview-page__readme mr-3">
                <ExtensionReadme extension={extension} history={history} />
            </div>
            <aside className="registry-extension-overview-page__sidebar">
                {categories && (
                    <div className="mb-3">
                        <RegistryExtensionOverviewIcon extension={extension} isLightTheme={isLightTheme} />
                        <h3>Categories</h3>
                        <ul className="list-inline test-registry-extension-categories">
                            {categories.map(category => (
                                <li key={category} className="list-inline-item mb-2">
                                    <Link
                                        to={urlToExtensionsQuery(extensionsQuery({ category }))}
                                        className="btn btn-outline-secondary btn-sm"
                                    >
                                        {category}
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
                {extension.manifest &&
                    !isErrorLike(extension.manifest) &&
                    extension.manifest.tags &&
                    extension.manifest.tags.length > 0 && (
                        <div className="mb-3">
                            <h3>Tags</h3>
                            <ul className="list-inline">
                                {extension.manifest.tags.map(tag => (
                                    <li key={tag} className="list-inline-item mb-2">
                                        <Link
                                            to={urlToExtensionsQuery(extensionsQuery({ tag }))}
                                            className="btn btn-outline-secondary btn-sm registry-extension-overview-page__tag"
                                        >
                                            {tag}
                                        </Link>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}
                <small className="text-muted">
                    <dl className="border-top pt-2">
                        {extension.registryExtension?.publisher && (
                            <>
                                <dt>Publisher</dt>
                                <dd>
                                    {extension.registryExtension.publisher ? (
                                        <Link to={extension.registryExtension.publisher.url}>
                                            {extensionIDPrefix(extension.registryExtension.publisher)}
                                        </Link>
                                    ) : (
                                        'Unavailable'
                                    )}
                                </dd>
                            </>
                        )}
                        {extension.registryExtension?.registryName && (
                            <>
                                <dt className={extension.registryExtension.publisher ? 'border-top pt-2' : ''}>
                                    Published on
                                </dt>
                                <dd>
                                    <span>{extension.registryExtension.registryName}</span>
                                </dd>
                            </>
                        )}
                        <dt className="border-top pt-2">Extension ID</dt>
                        <dd>{extension.id}</dd>
                        {extension.registryExtension &&
                            (extension.registryExtension.updatedAt || extension.registryExtension.publishedAt) && (
                                <>
                                    <dt className="border-top pt-2">Last updated</dt>
                                    <dd>
                                        <Timestamp
                                            date={maxDate(
                                                [
                                                    extension.registryExtension.updatedAt,
                                                    extension.registryExtension.publishedAt,
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
                            {extension.registryExtension && (
                                <Link to={`${extension.registryExtension.url}/-/manifest`} className="d-block">
                                    Manifest (package.json)
                                </Link>
                            )}
                            {extension.manifest && !isErrorLike(extension.manifest) && extension.manifest.url && (
                                <a
                                    href={extension.manifest.url}
                                    target="_blank"
                                    rel="nofollow noopener noreferrer"
                                    className="d-block"
                                >
                                    Source code (JavaScript)
                                </a>
                            )}
                            {repositoryURL && (
                                <div className="d-flex">
                                    {repositoryURL.hostname === 'github.com' && (
                                        <GithubIcon className="icon-inline mr-1" />
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

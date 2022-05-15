import React, { useMemo, useEffect } from 'react'

import classNames from 'classnames'
import { parseISO } from 'date-fns'
import maxDate from 'date-fns/max'
import { isObject } from 'lodash'
import GithubIcon from 'mdi-react/GithubIcon'

import { isErrorLike, isDefined, isEncodedImage } from '@sourcegraph/common'
import { splitExtensionID } from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionCategory, ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'
import { Button, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { DefaultExtensionIcon, DefaultSourcegraphExtensionIcon, SourcegraphExtensionIcon } from '../icons'

import { extensionsQuery, urlToExtensionsQuery, validCategories } from './extension'
import { ExtensionAreaRouteContext } from './ExtensionArea'
import { ExtensionReadme } from './RegistryExtensionReadme'
import { SourcegraphExtensionFeedback } from './SourcegraphExtensionFeedback'

import styles from './RegistryExtensionOverviewPage.module.scss'

interface Props extends Pick<ExtensionAreaRouteContext, 'extension' | 'telemetryService' | 'isLightTheme'> {}

const RegistryExtensionOverviewIcon: React.FunctionComponent<
    React.PropsWithChildren<Pick<Props, 'extension' | 'isLightTheme'>>
> = ({ extension, isLightTheme }) => {
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
        return <img className={classNames('mb-3', styles.icon)} src={iconURL.href} alt="" />
    }

    if (manifest?.publisher === 'sourcegraph') {
        return <DefaultSourcegraphExtensionIcon className={classNames('mb-3', styles.icon)} />
    }
    return <DefaultExtensionIcon className={classNames('mb-3', styles.icon)} />
}

/** A page that displays overview information about a registry extension. */
export const RegistryExtensionOverviewPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
    extension,
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

    const { publisher, isSourcegraphExtension } = splitExtensionID(extension.id)

    return (
        <div className="d-flex flex-wrap">
            <PageTitle title={extension.id} />
            <div className={classNames('mr-3', styles.readme)}>
                <ExtensionReadme extension={extension} />
            </div>
            <aside className={styles.sidebar}>
                <RegistryExtensionOverviewIcon extension={extension} isLightTheme={isLightTheme} />
                {/* Publisher */}
                {publisher && (
                    <div className="pt-2 pb-3">
                        <Typography.H3 as={Typography.H2}>Publisher</Typography.H3>
                        <small
                            data-tooltip={isSourcegraphExtension ? 'Created and maintained by Sourcegraph' : undefined}
                        >
                            {publisher}
                        </small>
                        {isSourcegraphExtension && <SourcegraphExtensionIcon className={styles.sourcegraphIcon} />}
                    </div>
                )}
                {/* Installs/user count will go here */}
                {/* Rating (readonly) will go here */}
                {/* Last updated */}
                {extension.registryExtension &&
                    (extension.registryExtension.updatedAt || extension.registryExtension.publishedAt) && (
                        <div className={styles.sidebarSection}>
                            <Typography.H3>Last updated</Typography.H3>
                            <small className="text-muted">
                                <Timestamp
                                    date={maxDate(
                                        [extension.registryExtension.updatedAt, extension.registryExtension.publishedAt]
                                            .filter(isDefined)
                                            .map(date => parseISO(date))
                                    )}
                                />
                            </small>
                        </div>
                    )}
                {/* Resources */}
                <div className={styles.sidebarSection}>
                    <Typography.H3>Resources</Typography.H3>
                    <small>
                        {extension.registryExtension && (
                            <Link to={`${extension.registryExtension.url}/-/manifest`} className="d-block mb-1">
                                Manifest (package.json)
                            </Link>
                        )}
                        {extension.manifest && !isErrorLike(extension.manifest) && extension.manifest.url && (
                            <Link
                                to={extension.manifest.url}
                                target="_blank"
                                rel="nofollow noopener noreferrer"
                                className="d-block mb-1"
                            >
                                Source code (JavaScript)
                            </Link>
                        )}
                        {repositoryURL && (
                            <div className="d-flex">
                                {repositoryURL.hostname === 'github.com' && <Icon className="mr-1" as={GithubIcon} />}
                                <Link
                                    to={repositoryURL.href}
                                    rel="nofollow noreferrer noopener"
                                    target="_blank"
                                    className="d-block"
                                >
                                    Repository
                                </Link>
                            </div>
                        )}
                    </small>
                </div>
                {/* Full extension ID */}
                <div className={styles.sidebarSection}>
                    <Typography.H3>Extension ID</Typography.H3>
                    <small className="text-muted">{extension.id}</small>
                </div>
                {/* Categories */}
                {categories && (
                    <div className={classNames('pb-0', styles.sidebarSection)}>
                        <Typography.H3>Categories</Typography.H3>
                        <ul className="list-inline" data-testid="test-registry-extension-categories">
                            {categories.map(category => (
                                <li key={category} className="list-inline-item mb-2">
                                    <Button
                                        to={urlToExtensionsQuery({ category })}
                                        variant="secondary"
                                        outline={true}
                                        size="sm"
                                        as={Link}
                                    >
                                        {category}
                                    </Button>
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
                {/* Tags */}
                {extension.manifest &&
                    !isErrorLike(extension.manifest) &&
                    extension.manifest.tags &&
                    extension.manifest.tags.length > 0 && (
                        <div className={classNames('pb-0', styles.sidebarSection)}>
                            <Typography.H3>Tags</Typography.H3>
                            <ul className="list-inline">
                                {extension.manifest.tags.map(tag => (
                                    <li key={tag} className="list-inline-item mb-2">
                                        <Button
                                            to={urlToExtensionsQuery({ query: extensionsQuery({ tag }) })}
                                            className={styles.tag}
                                            variant="secondary"
                                            outline={true}
                                            size="sm"
                                            as={Link}
                                        >
                                            {tag}
                                        </Button>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}
                {/* Rating widget will go here */}
                {/* Message the author */}
                {isSourcegraphExtension && (
                    <div className={styles.sidebarSection}>
                        <SourcegraphExtensionFeedback extensionID={extension.id} />
                    </div>
                )}
            </aside>
        </div>
    )
}

import React from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Typography } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../components/branding/BrandLogo'
import { PageRoutes } from '../../routes.constants'

import styles from './SearchPageFooter.module.scss'

const footerLinkSections: { name: string; links: { name: string; to: string; eventName?: string }[] }[] = [
    {
        name: 'Resources',
        links: [
            {
                name: 'Docs',
                to: 'https://docs.sourcegraph.com/',
            },
            { name: 'Learn', to: 'https://learn.sourcegraph.com/' },
            { name: 'Blog', to: 'https://about.sourcegraph.com/blog/' },
        ],
    },
    {
        name: 'Product',
        links: [
            { name: 'Changelog', to: 'https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md' },
            { name: 'Enterprise', to: 'https://about.sourcegraph.com/' },
            { name: 'Pricing', to: 'https://about.sourcegraph.com/pricing' },
        ],
    },
    {
        name: 'Company',
        links: [
            { name: 'About', to: 'https://about.sourcegraph.com/' },
            { name: 'Careers', to: 'https://about.sourcegraph.com/jobs/' },
            { name: 'Contact', to: 'https://about.sourcegraph.com/contact' },
        ],
    },
    {
        name: 'Integrate',
        links: [
            {
                name: 'Browser extensions',
                to: 'https://docs.sourcegraph.com/integration/browser_extension',
                eventName: 'BrowserExtensions',
            },
            {
                name: 'Editor plugins',
                to: 'https://docs.sourcegraph.com/integration/editor',
                eventName: 'EditorPlugins',
            },
            {
                name: 'Code host integrations',
                to: 'https://docs.sourcegraph.com/admin/external_service',
                eventName: 'CodeHostIntegrations',
            },
        ],
    },
]

export const SearchPageFooter: React.FunctionComponent<
    React.PropsWithChildren<ThemeProps & TelemetryProps & { isSourcegraphDotCom: boolean }>
> = ({ isLightTheme, telemetryService, isSourcegraphDotCom }) => {
    const assetsRoot = window.context?.assetsRoot || ''

    const logLinkClicked = (name: string): void => {
        telemetryService.log('HomepageFooterCTASelected', { name }, { name })
    }

    const logDevelopmentToolTimeClicked = (): void => {
        telemetryService.log('HomepageDevToolTimeClicked')
    }

    return isSourcegraphDotCom ? (
        <footer className={styles.footer}>
            <Link to={PageRoutes.Search} aria-label="Home" className="flex-shrink-0">
                <BrandLogo isLightTheme={isLightTheme} variant="symbol" className={styles.logo} />
            </Link>

            <ul className={classNames('d-flex flex-wrap list-unstyled', styles.mainList)}>
                {footerLinkSections.map(section => (
                    <li key={section.name} className={styles.linkSection}>
                        <Typography.H2 className={styles.linkSectionHeading}>{section.name}</Typography.H2>
                        <ul className="list-unstyled">
                            {section.links.map(link => (
                                <li key={link.name}>
                                    <Link
                                        to={link.to}
                                        onClick={() => logLinkClicked(link.eventName ?? link.name)}
                                        className={styles.link}
                                    >
                                        {link.name}
                                    </Link>
                                </li>
                            ))}
                        </ul>
                    </li>
                ))}
                <li>
                    <Link
                        to="https://info.sourcegraph.com/dev-tool-time"
                        className={styles.devToolTimeWrapper}
                        onClick={logDevelopmentToolTimeClicked}
                    >
                        <img
                            src={`${assetsRoot}/img/devtooltime-logo.svg`}
                            alt=""
                            className={styles.devToolTimeImage}
                        />
                        <div className={styles.devToolTimeText}>
                            <Typography.H2 className={styles.linkSectionHeading}>Dev tool time</Typography.H2>
                            <div>The show where developers talk about dev tools, productivity hacks, and more.</div>
                        </div>
                    </Link>
                </li>
            </ul>
        </footer>
    ) : (
        <footer className={classNames(styles.serverFooter, 'd-flex flex-column flex-lg-row align-items-center')}>
            <Typography.H4 as={Typography.H3} className="mb-2 mb-lg-0">
                Explore and extend
            </Typography.H4>
            <span className="d-flex flex-column flex-md-row align-items-center">
                <span className="d-flex flex-row mb-2 mb-md-0">
                    <Link
                        className="px-3"
                        to="https://docs.sourcegraph.com/integration/browser_extension"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={() => logLinkClicked('BrowserExtensions')}
                    >
                        Browser extensions
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                    <Link
                        className="px-3"
                        to="/extensions"
                        target="_blank"
                        onClick={() => logLinkClicked('SourcegraphExtensions')}
                    >
                        Sourcegraph extensions
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                </span>
                <span className="d-flex flex-row">
                    <Link
                        className="px-3"
                        to="https://docs.sourcegraph.com/integration/editor"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={() => logLinkClicked('EditorPlugins')}
                    >
                        Editor plugins
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                    <Link
                        className="pl-3"
                        to="https://docs.sourcegraph.com/admin/external_service"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={() => logLinkClicked('CodeHostIntegrations')}
                    >
                        Code host integrations
                    </Link>
                </span>
            </span>
        </footer>
    )
}

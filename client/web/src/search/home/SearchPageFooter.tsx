import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BrandLogo } from '../../components/branding/BrandLogo'

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
            { name: 'Careers', to: 'https://boards.greenhouse.io/sourcegraph91' },
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

export const SearchPageFooter: React.FunctionComponent<ThemeProps & TelemetryProps> = ({
    isLightTheme,
    telemetryService,
}) => {
    const assetsRoot = window.context?.assetsRoot || ''

    const logLinkClicked = (name: string): void => {
        telemetryService.log('HomepageFooterCTASelected', { name }, { name })
    }

    const logDevelopmentToolTimeClicked = (): void => {
        telemetryService.log('HomepageDevToolTimeClicked')
    }

    return (
        <footer className={styles.footer}>
            <Link to="/search" aria-label="Home" className="flex-shrink-0">
                <BrandLogo isLightTheme={isLightTheme} variant="symbol" className={styles.logo} />
            </Link>

            <ul className={classNames('d-flex flex-wrap list-unstyled', styles.mainList)}>
                {footerLinkSections.map(section => (
                    <li key={section.name} className={styles.linkSection}>
                        <h2 className={styles.linkSectionHeading}>{section.name}</h2>
                        <ul className="list-unstyled">
                            {section.links.map(link => (
                                <li key={link.name}>
                                    <a
                                        href={link.to}
                                        onClick={() => logLinkClicked(link.eventName ?? link.name)}
                                        className={styles.link}
                                    >
                                        {link.name}
                                    </a>
                                </li>
                            ))}
                        </ul>
                    </li>
                ))}
                <li>
                    <a
                        href="https://info.sourcegraph.com/dev-tool-time"
                        className={styles.devToolTimeWrapper}
                        onClick={logDevelopmentToolTimeClicked}
                    >
                        <img
                            src={`${assetsRoot}/img/devtooltime-logo.svg`}
                            alt="DevToolTime logo"
                            className={styles.devToolTimeImage}
                        />
                        <div className={styles.devToolTimeText}>
                            <h2 className={styles.linkSectionHeading}>Dev tool time</h2>
                            <div>The show where developers talk about dev tools, productivity hacks, and more.</div>
                        </div>
                    </a>
                </li>
            </ul>
        </footer>
    )
}

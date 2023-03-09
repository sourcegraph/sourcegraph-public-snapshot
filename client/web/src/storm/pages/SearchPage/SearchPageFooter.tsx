import { FC } from 'react'

import classNames from 'classnames'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Link, H2 } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../../components/branding/BrandLogo'
import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'
import { PageRoutes } from '../../../routes.constants'

import styles from './SearchPageFooter.module.scss'

export const SearchPageFooter: FC = () => {
    const isLightTheme = useIsLightTheme()
    const { telemetryService, isSourcegraphDotCom } = useLegacyContext_onlyInStormRoutes()

    const logLinkClicked = (name: string): void => {
        telemetryService.log('HomepageFooterCTASelected', { name }, { name })
    }

    const logDiscordClicked = (): void => {
        telemetryService.log('HomepageDiscordClicked')
    }

    if (isSourcegraphDotCom) {
        return (
            <footer className={styles.footer}>
                <Link to={PageRoutes.Search} aria-label="Home" className="flex-shrink-0">
                    <BrandLogo isLightTheme={isLightTheme} variant="symbol" className={styles.logo} />
                </Link>

                <ul className={classNames('d-flex flex-wrap list-unstyled', styles.mainList)}>
                    {footerLinkSections.map(section => (
                        <li key={section.name} className={styles.linkSection}>
                            <H2 className={styles.linkSectionHeading}>{section.name}</H2>
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
                        <div className={styles.discordText}>
                            <Link
                                to="https://srcgr.ph/discord-server"
                                className={styles.discordWrapper}
                                onClick={logDiscordClicked}
                            >
                                <H2 className={styles.linkSectionHeading}>Join our Discord</H2>
                            </Link>
                            <div>If you need help or want to share something with the community, join us!</div>
                        </div>
                    </li>
                </ul>
            </footer>
        )
    }

    return (
        <footer className={classNames(styles.serverFooter, 'd-flex flex-column flex-lg-row align-items-center')}>
            <span className="d-flex flex-column flex-md-row align-items-center">
                <span className="d-flex flex-row mb-2 mb-md-0">
                    <Link
                        className="px-3"
                        to="/help/integration/browser_extension"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={() => logLinkClicked('BrowserExtensions')}
                    >
                        Browser extensions
                    </Link>
                    <span aria-hidden="true" className="border-right d-none d-md-inline" />
                </span>
                <span className="d-flex flex-row">
                    <Link
                        className="px-3"
                        to="/help/integration/editor"
                        rel="noopener noreferrer"
                        target="_blank"
                        onClick={() => logLinkClicked('EditorPlugins')}
                    >
                        Editor plugins
                    </Link>
                </span>
            </span>
        </footer>
    )
}

interface FooterLinkSection {
    name: string
    links: { name: string; to: string; eventName?: string }[]
}

const footerLinkSections: FooterLinkSection[] = [
    {
        name: 'Resources',
        links: [
            {
                name: 'Docs',
                to: 'https://docs.sourcegraph.com/',
            },
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
                to: '/help/integration/browser_extension',
                eventName: 'BrowserExtensions',
            },
            {
                name: 'Editor plugins',
                to: '/help/integration/editor',
                eventName: 'EditorPlugins',
            },
        ],
    },
]

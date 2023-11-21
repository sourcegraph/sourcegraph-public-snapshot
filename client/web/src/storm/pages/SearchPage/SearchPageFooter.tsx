import { type FC, useMemo } from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'

import styles from './SearchPageFooter.module.scss'

export const SearchPageFooter: FC = () => {
    const { telemetryService, isSourcegraphDotCom } = useLegacyContext_onlyInStormRoutes()

    const logLinkClicked = (name: string): void => {
        telemetryService.log('HomepageFooterCTASelected', { name }, { name })
    }

    const links = useMemo(
        (): { name: string; href: string }[] =>
            isSourcegraphDotCom
                ? [
                      {
                          name: 'Docs',
                          href: 'https://docs.sourcegraph.com/',
                      },
                      { name: 'Homepage', href: 'https://sourcegraph.com' },
                      {
                          name: 'Cody',
                          href: 'https://sourcegraph.com/cody',
                      },
                      {
                          name: 'Enterprise',
                          href: 'https://sourcegraph.com/get-started?t=enterprise',
                      },
                      {
                          name: 'Security',
                          href: 'https://sourcegraph.com/security',
                      },
                      { name: 'Discord', href: 'https://srcgr.ph/discord-server' },
                  ]
                : [],
        [isSourcegraphDotCom]
    )

    return links.length === 0 ? null : (
        <footer className={classNames(styles.serverFooter, 'd-flex flex-column flex-lg-row align-items-center')}>
            <span className="d-flex flex-column flex-md-row align-items-center">
                {links.map(({ name, href }, index) => (
                    <span className="d-flex flex-row mb-2 mb-md-0" key={name}>
                        <Link
                            className="px-3 text-muted"
                            to={href}
                            rel="noopener noreferrer"
                            target="_blank"
                            onClick={() => logLinkClicked(name)}
                        >
                            {name}
                        </Link>
                        {index !== links.length - 1 && (
                            <span aria-hidden="true" className="border-right d-none d-md-inline" />
                        )}
                    </span>
                ))}
            </span>
        </footer>
    )
}

import { type FC, useMemo } from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'

import styles from './SearchPageFooter.module.scss'

type linkNames = 'Docs' | 'About' | 'Cody' | 'Enterprise' | 'Security' | 'Discord'

const v2LinkNameTypes: { [key in linkNames]: number } = {
    Docs: 1,
    About: 2,
    Cody: 3,
    Enterprise: 4,
    Security: 5,
    Discord: 6,
}

export const SearchPageFooter: FC = () => {
    const { telemetryService, isSourcegraphDotCom, platformContext } = useLegacyContext_onlyInStormRoutes()
    const { telemetryRecorder } = platformContext

    const logLinkClicked = (name: linkNames): void => {
        telemetryService.log('HomepageFooterCTASelected', { name }, { name })
        telemetryRecorder.recordEvent('home.footer.CTA', 'click', { metadata: { type: v2LinkNameTypes[name] } })
    }

    const links = useMemo(
        (): { name: linkNames; href: string }[] =>
            isSourcegraphDotCom
                ? [
                      {
                          name: 'Docs',
                          href: 'https://sourcegraph.com/docs',
                      },
                      { name: 'About', href: 'https://sourcegraph.com' },
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

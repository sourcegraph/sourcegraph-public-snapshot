import React from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link } from '@sourcegraph/wildcard'

import styles from './SelfHostedCtaLink.module.scss'

interface SelfHostedCtaLinkProps extends TelemetryProps {
    className?: string
    contentClassName?: string
    page: string
}

export const SelfHostedCtaLink: React.FunctionComponent<React.PropsWithChildren<SelfHostedCtaLinkProps>> = ({
    className,
    contentClassName,
    telemetryService,
    page,
}) => {
    const linkProps = { rel: 'noopener noreferrer' }

    const gettingStartedCTAOnClick = (): void => {
        telemetryService.log('InstallSourcegraphCTAClicked', { page }, { page })
    }

    return (
        <div className={classNames(styles.container, className)}>
            <span className={contentClassName}>
                Get Sourcegraph for teams via our{' '}
                <Link onClick={gettingStartedCTAOnClick} to="https://docs.sourcegraph.com/admin/install" {...linkProps}>
                    self-hosted installation
                </Link>
                .
            </span>
        </div>
    )
}

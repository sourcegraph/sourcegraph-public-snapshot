import classNames from 'classnames'
import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import styles from './SelfHostedCtaLink.module.scss'

interface SelfHostedCtaLinkProps extends TelemetryProps {
    className?: string
    contentClassName?: string
    page: string
}

export const SelfHostedCtaLink: React.FunctionComponent<SelfHostedCtaLinkProps> = ({
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
                <a onClick={gettingStartedCTAOnClick} href="https://docs.sourcegraph.com/admin/install" {...linkProps}>
                    self-hosted installation
                </a>
                .
            </span>
        </div>
    )
}

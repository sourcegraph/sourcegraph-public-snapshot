import React from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../components/DismissibleAlert'
import { PageRoutes } from '../routes.constants'
import { eventLogger, telemetryRecorder } from '../tracking/eventLogger'

const onClickCTA = (): void => {
    telemetryRecorder?.recordEvent('alertNeedsRepoConfigCTA', 'clicked')
    eventLogger.log('AlertNeedsRepoConfigCTAClicked')
}

/**
 * A global alert telling the site admin that they need to configure repositories
 * on this site.
 */
export const NeedsRepositoryConfigurationAlert: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string }>
> = ({ className }) => (
    <DismissibleAlert
        partialStorageKey="needsRepositoryConfiguration"
        variant="success"
        className={classNames('d-flex align-items-center', className)}
    >
        <Link className="site-alert__link" to={`${PageRoutes.SetupWizard}/remote-repositories`} onClick={onClickCTA}>
            <span className="underline">Go to setup wizard</span>
        </Link>
        &nbsp;to add remote repositories from GitHub, GitLab, etc.
    </DismissibleAlert>
)

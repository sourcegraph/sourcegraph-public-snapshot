import Icon from '@sourcegraph/icons/lib/Download'
import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickCTA = () => {
    eventLogger.log('AlertUpdateAvailableCTAClicked')
}

/**
 * A global alert telling the site admin that an updated version of the Sourcegraph
 * Docker image is available.
 */
export const UpdateAvailableAlert: React.SFC<{
    updateVersionAvailable: string
    className?: string
}> = ({ updateVersionAvailable, className = '' }) => (
    <DismissibleAlert
        partialStorageKey={`Update/${updateVersionAvailable}`}
        className={`alert-animated-bg alert-success update-available-alert ${className}`}
    >
        <Icon className="icon-inline site-alert__link-icon" /> An update is available:&nbsp;
        <a className="site-alert__link" href="https://about.sourcegraph.com" onClick={onClickCTA}>
            <span className="underline">Sourcegraph Server {updateVersionAvailable}</span>
        </a>
    </DismissibleAlert>
)

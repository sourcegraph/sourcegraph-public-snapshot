import Icon from '@sourcegraph/icons/lib/Download'
import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'
import { DismissibleAlert } from './DismissibleAlert'

const onClickCTA = () => {
    eventLogger.log('AlertUpdateAvailableCTAClicked')
}

/**
 * A global alert telling the site admin that an updated version of the Sourcegraph
 * Docker image is available.
 */
export const UpdateAvailableAlert: React.SFC<{ updateVersionAvailable: string }> = ({ updateVersionAvailable }) => (
    <DismissibleAlert
        partialStorageKey={`Update/${updateVersionAvailable}`}
        className="site-alert alert-success update-available-alert"
    >
        <Icon className="icon-inline site-alert__link-icon" /> An update is available:&nbsp;
        <a className="site-alert__link" href="https://about.sourcegraph.com" onClick={onClickCTA}>
            <span className="underline">Sourcegraph Server {updateVersionAvailable}</span>
        </a>
    </DismissibleAlert>
)

import CloudDownloadIcon from 'mdi-react/CloudDownloadIcon'
import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickCTA = () => {
    eventLogger.log('AlertUpdateAvailableCTAClicked')
}
const onClickChangelog = () => {
    eventLogger.log('AlertUpdateAvailableChangelogClicked')
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
        <CloudDownloadIcon className="icon-inline site-alert__link-icon mr-2" />An update is available:&nbsp;
        <a className="site-alert__link" href="https://about.sourcegraph.com" target="_blank" onClick={onClickCTA}>
            <span className="underline">Sourcegraph {updateVersionAvailable}</span>
        </a>&nbsp;-&nbsp;
        <a
            className="site-alert__link"
            href="https://about.sourcegraph.com/changelog"
            target="_blank"
            onClick={onClickChangelog}
        >
            changelog
        </a>
    </DismissibleAlert>
)

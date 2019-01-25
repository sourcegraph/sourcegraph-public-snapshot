import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickInstall = (): void => {
    eventLogger.log('InstallSourcegraphServerCTAClicked', { location_on_page: 'banner' })
}

export const ServerBanner = () => (
    <DismissibleAlert partialStorageKey="set-up-self-hosted" className="alert alert-info">
        <span>
            Search your private and internal code.{' '}
            <a href="https://docs.sourcegraph.com/#quickstart" onClick={onClickInstall}>
                Set up a self-hosted Sourcegraph instance.
            </a>
        </span>
    </DismissibleAlert>
)

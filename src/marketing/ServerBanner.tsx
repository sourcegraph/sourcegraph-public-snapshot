import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'

const onClickInstall = (): void => {
    eventLogger.log('InstallSourcegraphServerCTAClicked', { location_on_page: 'banner' })
}

export const ServerBanner = () => (
    <div className="alert alert-secondary">
        Search your private and internal code.{' '}
        <a href="https://about.sourcegraph.com" onClick={onClickInstall}>
            Set up a self-hosted Sourcegraph instance.
        </a>
    </div>
)

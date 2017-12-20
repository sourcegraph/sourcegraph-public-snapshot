import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'
import { Banner } from './Banner'

const onClickInstall = (): void => {
    eventLogger.log('InstallSourcegraphServerCTAClicked', { location_on_page: 'banner' })
}

export const ServerBanner = () => (
    <Banner
        title="Search your private and internal code."
        ctaLink="https://about.sourcegraph.com"
        ctaText="Install Sourcegraph Server"
        onClick={onClickInstall}
    />
)

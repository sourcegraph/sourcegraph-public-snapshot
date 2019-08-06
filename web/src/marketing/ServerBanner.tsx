import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickInstall = (): void => {
    eventLogger.log('InstallSourcegraphServerCTAClicked', { location_on_page: 'banner' })
}

export const ServerBanner: React.FunctionComponent = () => (
    <DismissibleAlert partialStorageKey="set-up-self-hosted" className="alert alert-info">
        <span>
            Sourcegraph.com searches over the top 500 GitHub repositories by default. You can search over other public
            repositories by providing a repo: filter or you can search all of your own private code by{' '}
            <a href="https://docs.sourcegraph.com/#quickstart" onClick={onClickInstall}>
                setting up a self-hosted Sourcegraph instance.
            </a>
        </span>
    </DismissibleAlert>
)

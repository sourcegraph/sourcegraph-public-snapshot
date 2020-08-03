import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickInstall = (): void => {
    eventLogger.log('InstallSourcegraphServerCTAClicked')
}

export const ServerBanner: React.FunctionComponent = () => (
    <DismissibleAlert partialStorageKey="set-up-self-hosted" className="alert alert-info">
        <span>
            Search your private and internal code.{' '}
            <a href="https://docs.sourcegraph.com/#quickstart" onClick={onClickInstall}>
                Set up a self-hosted Sourcegraph instance.
            </a>
        </span>
    </DismissibleAlert>
)

export const ServerBannerNoRepo: React.FunctionComponent = () => (
    <DismissibleAlert partialStorageKey="set-up-self-hosted" className="alert alert-info">
        <span>
            If an open-source repository is not yet indexed, you can add it by going to <code>sourcegraph.com/github.com/$OWNER/$REPO</code>,
            or search your private code by
            <a href="https://docs.sourcegraph.com/#quickstart" onClick={onClickInstall}>
                setting up a self-hosted Sourcegraph instance.
            </a>
        </span>
    </DismissibleAlert>
)

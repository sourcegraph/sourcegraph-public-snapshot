import WarningIcon from 'mdi-react/WarningIcon'
import * as React from 'react'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickCTA = (): void => {
    eventLogger.log('AlertPerformanceWarningCTAClicked')
}

/**
 * An alert that explains the performance limitations of single-node Docker deployments.
 */
export const PerformanceWarningAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        partialStorageKey="performanceWarningAlert"
        className="alert alert-warning align-items-center m-2"
    >
        <WarningIcon className="icon-inline mr-2 flex-shrink-0" />
        <div>
            Search performance and accuracy are limited on single-node Docker deployments. We recommend that instances
            with 100+ repositories&nbsp;
            <a
                target="_blank"
                rel="noopener"
                className="site-alert__link"
                href="https://docs.sourcegraph.com/admin/install"
                onClick={onClickCTA}
            >
                deploy to a cluster
            </a>
            &nbsp;for optimal performance.&nbsp;
            <a target="_blank" rel="noopener" className="site-alert__link" href="https://about.sourcegraph.com/contact">
                Contact us
            </a>
            &nbsp;for support or to learn more.
        </div>
    </DismissibleAlert>
)

import * as React from 'react'
import { Link } from 'react-router-dom'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'
import { CircleChevronRightIcon } from '../util/icons' // TODO: Switch to mdi icon

const onClickCTA = () => {
    eventLogger.log('AlertNeedsRepoConfigCTAClicked')
}

/**
 * A global alert telling the site admin that they need to configure repositories
 * on this site.
 */
export const NeedsRepositoryConfigurationAlert: React.SFC<{ className?: string }> = ({ className = '' }) => (
    <DismissibleAlert
        partialStorageKey="needsRepositoryConfiguration"
        className={`alert alert-success alert-animated-bg d-flex align-items-center ${className}`}
    >
        <Link className="site-alert__link" to="/site-admin/configuration" onClick={onClickCTA}>
            <CircleChevronRightIcon className="icon-inline site-alert__link-icon" />{' '}
            <span className="underline">Configure repositories and code hosts</span>
        </Link>
        &nbsp;to add to Sourcegraph.
    </DismissibleAlert>
)

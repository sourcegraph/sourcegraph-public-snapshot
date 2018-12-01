import * as React from 'react'
import { Link } from 'react-router-dom'
import { CircleChevronRightIcon } from '../../../shared/src/components/icons' // TODO: Switch to mdi icon
import { DismissibleAlert } from '../components/DismissibleAlert'
import { eventLogger } from '../tracking/eventLogger'

const onClickCTA = () => {
    eventLogger.log('AlertNoReposEnabledCTAClicked')
}

/**
 * A global alert telling the site admin that they need to enable access to repositories
 * on this site.
 */
export const NoRepositoriesEnabledAlert: React.FunctionComponent<{ className?: string }> = ({ className = '' }) => (
    <DismissibleAlert
        partialStorageKey="noRepositoriesEnabled"
        className={`alert alert-success alert-animated-bg d-flex align-items-center ${className}`}
    >
        <Link className="site-alert__link" to="/site-admin/repositories" onClick={onClickCTA}>
            <CircleChevronRightIcon className="icon-inline site-alert__link-icon" />{' '}
            <span className="underline"> Select repositories to enable</span>
        </Link>
        &nbsp;to start searching and browsing
    </DismissibleAlert>
)

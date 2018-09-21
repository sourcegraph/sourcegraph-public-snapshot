import * as React from 'react'
import { Link } from 'react-router-dom'
import { eventLogger } from '../tracking/eventLogger'
import { CircleChevronRightIcon } from '../util/icons' // TODO: Switch to mdi icon

const onClickCTA = () => {
    eventLogger.log('AlertNoReposEnabledCTAClicked')
}

/**
 * A global alert telling the site admin that they need to enable access to repositories
 * on this site.
 */
export const NoRepositoriesEnabledAlert: React.SFC<{ className?: string }> = ({ className = '' }) => (
    <div className={`alert alert-success alert-animated-bg d-flex align-items-center ${className}`}>
        <Link className="site-alert__link" to="/site-admin/repositories" onClick={onClickCTA}>
            <CircleChevronRightIcon className="icon-inline site-alert__link-icon" />{' '}
            <span className="underline"> Select repositories to enable</span>
        </Link>
        &nbsp;to start searching and browsing
    </div>
)

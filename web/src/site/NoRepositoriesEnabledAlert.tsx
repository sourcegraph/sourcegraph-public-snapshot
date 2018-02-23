import Icon from '@sourcegraph/icons/lib/CircleChevronRight'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { eventLogger } from '../tracking/eventLogger'

const onClickCTA = () => {
    eventLogger.log('AlertNoReposEnabledCTAClicked')
}

/**
 * A global alert telling the site admin that they need to enable access to repositories
 * on this site.
 */
export const NoRepositoriesEnabledAlert: React.SFC = () => (
    <div className="alert alert-success site-alert no-repositories-enabled-alert">
        <Link className="site-alert__link" to="/site-admin/repositories" onClick={onClickCTA}>
            <Icon className="icon-inline site-alert__link-icon" />{' '}
            <span className="underline">Select repositories to enable</span>
        </Link>
        &nbsp;to start searching and browsing
    </div>
)

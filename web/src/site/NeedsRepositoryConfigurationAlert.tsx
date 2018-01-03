import Icon from '@sourcegraph/icons/lib/CircleChevronRight'
import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A global alert telling the site admin that they need to configure repositories
 * on this site.
 */
export const NeedsRepositoryConfigurationAlert: React.SFC = () => (
    <div className="alert alert-success site-alert needs-repository-configuration-alert">
        <Link className="site-alert__link" to="/site-admin/configuration">
            <Icon className="icon-inline site-alert__link-icon" />{' '}
            <span className="underline">Configure repositories and code hosts</span>
        </Link>
        &nbsp;to add to Sourcegraph Server.
    </div>
)

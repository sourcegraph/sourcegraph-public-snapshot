import React from 'react'

import classNames from 'classnames'

import { Link } from '@sourcegraph/wildcard'

import { DismissibleAlert } from '../components/DismissibleAlert'
import { PageRoutes } from '../routes.constants'
import { eventLogger } from '../tracking/eventLogger'

const onClickCTA = (): void => {
    eventLogger.log('AlertNeedsRepoConfigCTAClicked')
}

/**
 * A global alert telling the site admin that they need to configure repositories
 * on this site.
 */
export const NeedsRepositoryConfigurationAlert: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string }>
> = ({ className }) => (
    <DismissibleAlert
        partialStorageKey="needsRepositoryConfiguration"
        variant="success"
        className={classNames('d-flex align-items-center', className)}
    >
        <Link className="site-alert__link" to={PageRoutes.SetupWizard} onClick={onClickCTA}>
            <span className="underline">Connect a code host</span>
        </Link>
        &nbsp;to connect repositories to Sourcegraph.
    </DismissibleAlert>
)

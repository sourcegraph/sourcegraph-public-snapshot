import classNames from 'classnames'
import H from 'history'
import React, { useEffect } from 'react'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { siteAdminOverviewComponents } from './overviewComponents'

interface Props extends ActivationProps {
    history: H.History
    overviewComponents: typeof siteAdminOverviewComponents
}

/**
 * A page displaying an overview of site admin information.
 */
export const SiteAdminOverviewPage: React.FunctionComponent<Props> = ({ history, overviewComponents, activation }) => {
    useEffect(() => {
        eventLogger.logViewEvent('SiteAdminOverview')
    }, [])

    return (
        <div className="site-admin-overview-page">
            <PageTitle title="Overview - Admin" />
            <div className="site-admin-overview-page__grid">
                {overviewComponents.map(({ component: C, noCardClass, fullWidth }, i) => (
                    <div
                        className={classNames('site-admin-overview-page__grid-cell', {
                            'site-admin-overview-page__grid-cell--full-width': fullWidth,
                        })}
                        // This array index is statically defined, so it is stable.
                        //
                        // eslint-disable-next-line react/no-array-index-key
                        key={i}
                    >
                        <div className={noCardClass ? '' : 'card'}>
                            <C activation={activation} history={history} />
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}

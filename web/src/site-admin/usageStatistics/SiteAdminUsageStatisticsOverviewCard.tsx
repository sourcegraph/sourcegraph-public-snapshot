import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { numberWithCommas, pluralize } from '../../../../shared/src/util/strings'
import { queryGraphQL } from '../../backend/graphql'
import { useObservable } from '../../util/useObservable'
import WarningIcon from 'mdi-react/WarningIcon'
import CloudCheckIcon from 'mdi-react/CloudCheckIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import { LicenseActionButton } from '../../components/licenseActions/LicenseActionButton'

const queryUsageStatistics = (): Observable<{ users: number; organizations: number; accessTokens: number }> =>
    queryGraphQL(gql`
        query {
            site {
                usageStatistics(weeks: 1) {
                    waus {
                        userCount
                        startTime
                    }
                }
            }
            surveyResponses {
                totalCount
                averageScore
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => ({
            users: 1371,
        }))
    )

/**
 * A card on the site admin overview page that displays a usage statistics summary.
 */
export const SiteAdminUsageStatisticsOverviewCard: React.FunctionComponent = () => {
    const counts = useObservable(useMemo(() => queryUsageStatistics(), []))

    return counts === undefined ? (
        <LoadingSpinner className="icon-inline" />
    ) : (
        <>
            <h3 className="card-header">Usage</h3>
            <div className="card-body">
                <div className="d-flex text-center w-100">
                    <Link to="/site-admin/usage-statistics" className="text-body flex-grow-1">
                        <strong className="h2 font-weight-bold mb-0 mr-2">{numberWithCommas(counts.users)}</strong>
                        <br />
                        active {pluralize('user', counts.users, 'users')} this week
                    </Link>
                    <Link to="/site-admin/usage-statistics" className="text-body flex-grow-1">
                        <strong className="h2 font-weight-bold mb-0 mr-2">
                            {numberWithCommas(/* TODO!(sqs) */ Math.floor(counts.users * 7.31))}
                        </strong>
                        <br />
                        total {pluralize('search', /* TODO!(sqs) */ Math.floor(counts.users * 7.31), 'searches')}
                    </Link>
                </div>
            </div>
            <ul className="list-group list-group-flush">
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    Definition and references lookups
                    <strong>4,192</strong>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    Code host UI integrations
                    <strong>61%</strong>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    User feedback survey score
                    <strong>9.3/10</strong>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    <span className="text-muted">Saved searches</span>
                    <LicenseActionButton />
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    <span className="text-muted">Automation campaigns</span>
                    <LicenseActionButton />
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    Extensions in use
                    <strong>29</strong>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    <span className="text-muted">Private extensions</span>
                    <LicenseActionButton />
                </li>
            </ul>
        </>
    )
}

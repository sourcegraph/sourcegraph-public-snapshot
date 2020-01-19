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

const queryTotalCounts = (): Observable<{ users: number; organizations: number; accessTokens: number }> =>
    queryGraphQL(gql`
        query {
            users {
                totalCount
            }
            organizations {
                totalCount
            }
            site {
                accessTokens {
                    totalCount
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => ({
            users: data.users.totalCount,
            organizations: data.organizations.totalCount,
            accessTokens: data.site.accessTokens.totalCount,
        }))
    )

/**
 * A card on the site admin overview page that displays a user summary.
 */
export const SiteAdminUsersOverviewCard: React.FunctionComponent = () => {
    const counts = useObservable(useMemo(() => queryTotalCounts(), []))

    return counts === undefined ? (
        <LoadingSpinner className="icon-inline" />
    ) : (
        <>
            <h3 className="card-header">
                <Link to="/site-admin/users" className="text-body">
                    {numberWithCommas(counts.users)} {pluralize('user', counts.users)}
                </Link>
            </h3>
            <ul className="list-group list-group-flush">
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    Site admins <span className="font-weight-bold">7</span>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    Organizations <span className="font-weight-bold">{counts.organizations}</span>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    Access tokens <span className="font-weight-bold">{counts.accessTokens}</span>
                </li>
                <li className="list-group-item">
                    <div className="d-flex justify-content-between align-items-center">
                        Authentication providers <span className="font-weight-bold">2</span>
                    </div>
                    <ul className="list-inline">
                        <li className="list-inline-item">
                            <span className="badge badge-secondary font-weight-normal">Builtin username-password</span>
                        </li>
                        <li className="list-inline-item">
                            <span className="badge badge-secondary font-weight-normal">Google OAuth</span>
                        </li>
                    </ul>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    <span>
                        <WarningIcon className="icon-inline text-warning" /> 51 users have no verified email address
                    </span>
                    <button type="button" className="btn btn-sm btn-secondary">
                        Help
                    </button>
                </li>
            </ul>
            <footer className="card-footer">
                <ul className="list-inline mb-0">
                    <li className="list-inline-item">
                        <button type="button" className="btn btn-sm btn-secondary">
                            Create user
                        </button>
                    </li>
                    <li className="list-inline-item">
                        <button type="button" className="btn btn-sm btn-secondary">
                            Manage users
                        </button>
                    </li>
                </ul>
            </footer>
        </>
    )
}

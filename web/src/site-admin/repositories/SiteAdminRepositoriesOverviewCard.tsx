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
import { LicenseActionButton } from '../../components/licenseActions/LicenseActionButton'
import CloudAlertIcon from 'mdi-react/CloudAlertIcon'

const queryRepositoriesTotalCount = (): Observable<number> =>
    queryGraphQL(gql`
        query {
            repositories {
                totalCount(precise: true)
            }
        }
    `).pipe(
        map(dataOrThrowErrors),

        // Can't be null because the query requests a precise count.
        //
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        map(data => data.repositories.totalCount!)
    )

/**
 * A card on the site admin overview page that displays a repository summary.
 */
export const SiteAdminRepositoriesOverviewCard: React.FunctionComponent = () => {
    const repositoriesCount = useObservable(useMemo(() => queryRepositoriesTotalCount(), []))

    return repositoriesCount === undefined ? (
        <LoadingSpinner className="icon-inline" />
    ) : (
        <>
            <h3 className="card-header">
                <Link to="/site-admin/repositories" className="text-body">
                    {numberWithCommas(repositoriesCount)} {pluralize('repository', repositoriesCount, 'repositories')}
                </Link>
            </h3>
            <ul className="list-group list-group-flush">
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    GitLab <span className="font-weight-bold">19</span>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    GitHub Enterprise <span className="font-weight-bold">341</span>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    GitHub.com <span className="font-weight-bold">17</span>
                </li>
                <li className="list-group-item">
                    <CloudCheckIcon className="icon-inline text-success" /> All repositories are up to date{' '}
                    <span className="text-muted">as of 3 minutes ago</span>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-start">
                    <div className="d-flex">
                        <CloudAlertIcon className="icon-inline text-warning mr-1" />
                        <span>
                            Immediate repository updates disabled
                            <br />
                            <span className="text-muted">Last update: 3 days ago</span>
                            <br />
                            <span className="text-muted">Next update: 4 days from now</span>
                        </span>
                    </div>
                    <LicenseActionButton />
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    <span>
                        <WarningIcon className="icon-inline text-warning" /> GitLab UI integration not enabled
                    </span>
                    <button type="button" className="btn btn-sm btn-secondary">
                        Configure
                    </button>
                </li>
                <li className="list-group-item d-flex justify-content-between align-items-center">
                    <span>
                        <WarningIcon className="icon-inline text-warning" /> Repository permissions not enforced
                    </span>
                    <LicenseActionButton />
                </li>
            </ul>
            <footer className="card-footer">
                <ul className="list-inline mb-0">
                    <li className="list-inline-item">
                        <button type="button" className="btn btn-sm btn-secondary">
                            Manage repositories
                        </button>
                    </li>
                </ul>
            </footer>
        </>
    )
}

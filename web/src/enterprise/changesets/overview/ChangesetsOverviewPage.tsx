import React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { PageTitle } from '../../../components/PageTitle'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { threadsQueryWithValues } from '../../threads/url'
import { ChangesetsAreaTitle } from '../components/ChangesetsAreaTitle'
import { ChangesetsAreaContext } from '../global/ChangesetsArea'
import { ChangesetThreadsList } from '../threads/list/ChangesetThreadsList'

interface Props extends ChangesetsAreaContext, RouteComponentProps<{}> {}

/**
 * The changesets overview page.
 */
export const ChangesetsOverviewPage: React.FunctionComponent<Props> = ({ authenticatedUser, match, ...props }) => (
    <div className="changesets-overview-page container mt-3">
        <PageTitle title="Changesets" />
        <ChangesetsAreaTitle
            primaryActions={
                <Link to={`${match.url}/new`} className="btn btn-success">
                    New changeset
                </Link>
            }
        />
        <WithQueryParameter
            defaultQuery={threadsQueryWithValues('', {
                is: [props.type.toLowerCase(), 'open'],
                involves: authenticatedUser ? [authenticatedUser.username] : null,
            })}
            history={props.history}
            location={props.location}
        >
            {({ query, onQueryChange }) => (
                <ChangesetThreadsList
                    {...props}
                    authenticatedUser={authenticatedUser}
                    query={query}
                    onQueryChange={onQueryChange}
                />
            )}
        </WithQueryParameter>
    </div>
)

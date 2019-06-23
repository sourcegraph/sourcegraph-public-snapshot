import React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { PageTitle } from '../../../components/PageTitle'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { threadsQueryWithValues } from '../../threads/url'
import { ChecksAreaTitle } from '../components/ChecksAreaTitle'
import { ChecksAreaContext } from '../global/ChecksArea'
import { CheckThreadsList } from '../threads/list/CheckThreadsList'

interface Props extends ChecksAreaContext, RouteComponentProps<{}> {}

/**
 * The checks overview page.
 */
export const ChecksOverviewPage: React.FunctionComponent<Props> = ({ authenticatedUser, match, ...props }) => (
    <div className="checks-overview-page container mt-3">
        <PageTitle title="Checks" />
        <ChecksAreaTitle
            primaryActions={
                <Link to={`${match.url}/new`} className="btn btn-success">
                    New check
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
                <CheckThreadsList
                    {...props}
                    authenticatedUser={authenticatedUser}
                    query={query}
                    onQueryChange={onQueryChange}
                />
            )}
        </WithQueryParameter>
    </div>
)

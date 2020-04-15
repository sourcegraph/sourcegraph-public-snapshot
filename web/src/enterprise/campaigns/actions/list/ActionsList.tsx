import React from 'react'
import { queryActions } from '../backend'
import { Link } from '../../../../../../shared/src/components/Link'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { IUser, IAction } from '../../../../../../shared/src/graphql/schema'
import { ActionNode } from './ActionNode'

interface Props extends Pick<RouteComponentProps, 'history' | 'location'> {
    authenticatedUser: IUser
}

/**
 * A list of all actions
 */
export const ActionsList: React.FunctionComponent<Props> = props => (
    <>
        <h1>Actions</h1>
        <p>Run and track large-scale code changes on your machines</p>
        <Link to="/campaigns" className="btn btn-primary mb-3">
            Manage campaigns
        </Link>
        <Link to="/campaigns/actions/new" className="btn btn-primary ml-2 mb-3">
            New action
        </Link>

        <FilteredConnection<Pick<IAction, 'id' | 'name' | 'savedSearch' | 'schedule' | 'actionExecutions'>>
            {...props}
            nodeComponent={ActionNode}
            queryConnection={queryActions}
            hideSearch={true}
            noun="action"
            pluralNoun="actions"
        />
    </>
)

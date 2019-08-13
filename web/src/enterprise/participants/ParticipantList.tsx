import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ConnectionListHeader, ConnectionListHeaderItems } from '../../components/connectionList/ConnectionListHeader'
import { QueryParameterProps } from '../../components/withQueryParameter/WithQueryParameter'
import { ParticipantListItem, ParticipantListItemContext } from './ParticipantListItem'

const LOADING: 'loading' = 'loading'

export interface ParticipantListContext extends ParticipantListItemContext {
    /**
     * Whether each item should have a checkbox.
     */
    itemCheckboxes?: boolean

    history: H.History
    location: H.Location
}

interface Props extends QueryParameterProps, ParticipantListContext {
    participants: typeof LOADING | GQL.IParticipantConnection | ErrorLike

    headerItems?: ConnectionListHeaderItems

    className?: string
}

/**
 * A list of participants.
 */
export const ParticipantList: React.FunctionComponent<Props> = ({
    participants,
    headerItems,
    itemCheckboxes,
    className = '',
    ...props
}) => (
    <div className={`participant-list ${className}`}>
        {isErrorLike(participants) ? (
            <div className="alert alert-danger">{participants.message}</div>
        ) : (
            <div className="card">
                <ConnectionListHeader {...props} items={headerItems} itemCheckboxes={itemCheckboxes} />
                {participants === LOADING ? (
                    <LoadingSpinner className="m-3" />
                ) : participants.edges.length === 0 ? (
                    <p className="p-2 mb-0 text-muted">No participants.</p>
                ) : (
                    <ul className="list-group list-group-flush">
                        {participants.edges.map((edge, i) => (
                            <ParticipantListItem key={i} {...props} participant={edge} />
                        ))}
                    </ul>
                )}
            </div>
        )}
    </div>
)

export const ParticipantListHeaderCommonFilters: React.FunctionComponent<ParticipantListFilterContext> = props => (
    <>
        <ThreadListRepositoryFilterDropdownButton {...props} />
        <ThreadListLabelFilterDropdownButton {...props} />
    </>
)

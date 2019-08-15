import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useCallback, useState } from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ConnectionListFilterQueryInput } from '../../components/connectionList/ConnectionListFilterQueryInput'
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
    query,
    onQueryChange,
    className = '',
    ...props
}) => {
    const DEFAULT_MAX = 12
    const [max, setMax] = useState<number | undefined>(DEFAULT_MAX)
    const onShowAllClick = useCallback(() => setMax(undefined), [])

    return (
        <div className={`participant-list ${className}`}>
            {isErrorLike(participants) ? (
                <div className="alert alert-danger">{participants.message}</div>
            ) : (
                <div className="card">
                    <ConnectionListHeader
                        {...props}
                        items={{
                            ...headerItems,
                            left: (
                                <>
                                    <h4 className="mb-0">Participants</h4>
                                    {headerItems && headerItems.left}
                                </>
                            ),
                            right: (
                                <>
                                    {headerItems && headerItems.right}
                                    <ConnectionListFilterQueryInput
                                        query={query}
                                        onQueryChange={onQueryChange}
                                        instant={true}
                                    />
                                </>
                            ),
                        }}
                        itemCheckboxes={itemCheckboxes}
                    />
                    {participants === LOADING ? (
                        <LoadingSpinner className="m-3" />
                    ) : participants.edges.length === 0 ? (
                        <p className="p-3 mb-0 text-muted">No participants.</p>
                    ) : (
                        <>
                            <ul className="list-group list-group-flush">
                                {participants.edges.slice(0, max).map((edge, i) => (
                                    <ParticipantListItem key={i} {...props} participant={edge} />
                                ))}
                            </ul>
                            {max !== undefined && participants.edges.length > max && (
                                <div className="card-footer p-0">
                                    <button type="button" className="btn btn-sm btn-link" onClick={onShowAllClick}>
                                        Show {participants.edges.length - max} more
                                    </button>
                                </div>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    )
}

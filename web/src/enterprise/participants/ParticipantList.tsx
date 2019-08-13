import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ParticipantListItem, ParticipantListItemContext } from './ParticipantListItem'

const LOADING: 'loading' = 'loading'

interface Props extends ParticipantListItemContext {
    participants: typeof LOADING | GQL.IParticipantConnection | ErrorLike

    className?: string
}

/**
 * A list of participants.
 */
export const ParticipantList: React.FunctionComponent<Props> = ({ participants, className = '', ...props }) => (
    <div className={`participant-list ${className}`}>
        {isErrorLike(participants) ? (
            <div className="alert alert-danger mt-2">{participants.message}</div>
        ) : (
            <div className="card">
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

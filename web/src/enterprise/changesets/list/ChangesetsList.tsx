import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ChangesetListItem } from './ChangesetListItem'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerNotificationProps {
    changesets: typeof LOADING | GQL.IChangesetConnection | ErrorLike
}

/**
 * Lists changesets.
 */
export const ChangesetsList: React.FunctionComponent<Props> = ({ changesets, ...props }) => (
    <div className="changesets-list">
        {changesets === LOADING ? (
            <LoadingSpinner className="icon-inline mt-3" />
        ) : isErrorLike(changesets) ? (
            <div className="alert alert-danger mt-3">{changesets.message}</div>
        ) : (
            <div className="card">
                <div className="card-header">
                    <span className="text-muted">
                        {changesets.totalCount} {pluralize('changeset', changesets.totalCount)}
                    </span>
                </div>
                {changesets.nodes.length > 0 ? (
                    <ul className="list-group list-group-flush">
                        {changesets.nodes.map(changeset => (
                            <li key={changeset.id} className="list-group-item">
                                <ChangesetListItem {...props} changeset={changeset} />
                            </li>
                        ))}
                    </ul>
                ) : (
                    <div className="p-2 text-muted">No changesets yet.</div>
                )}
            </div>
        )}
    </div>
)

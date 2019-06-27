import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { WithChecklistsQueryResults } from '../components/withChecklistQueryResults/WithChecklistsQueryResults'
import { ChecklistsAreaContext } from '../global/ChecklistsArea'
import { ChecklistsListItem } from './item/ChecklistsListItem'

export interface ChecklistsListContext {
    itemClassName?: string
}

interface Props
    extends Partial<QueryParameterProps>,
        ChecklistsListContext,
        ChecklistsAreaContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The list of checklists with a header.
 */
export const ChecklistsList: React.FunctionComponent<Props> = ({ itemClassName, query, ...props }) => (
    <WithChecklistsQueryResults {...props} query={query}>
        {({ checklistsOrError }) => (
            <div className="checklists-list">
                {isErrorLike(checklistsOrError) ? (
                    <div className={itemClassName}>
                        <div className="alert alert-danger mt-2">{checklistsOrError.message}</div>
                    </div>
                ) : checklistsOrError === LOADING ? (
                    <div className={itemClassName}>
                        <LoadingSpinner className="mt-3" />
                    </div>
                ) : checklistsOrError.length === 0 ? (
                    <div className={itemClassName}>
                        <p className="p-2 mb-0 text-muted">No checklists found.</p>
                    </div>
                ) : (
                    <ul className="list-group list-group-flush mb-0">
                        {checklistsOrError.map((checklist, i) => (
                            <li key={i} className="list-group-item px-0">
                                <ChecklistsListItem
                                    {...props}
                                    key={JSON.stringify(checklist)}
                                    diagnostic={checklist}
                                    className={itemClassName}
                                />
                            </li>
                        ))}
                    </ul>
                )}
                <style>
                    {/* HACK TODO!(sqs) */}
                    {
                        '.checklists-list .markdown pre,.checklists-list .markdown code {margin:0;padding:0;background-color:transparent;}'
                    }
                </style>
            </div>
        )}
    </WithChecklistsQueryResults>
)

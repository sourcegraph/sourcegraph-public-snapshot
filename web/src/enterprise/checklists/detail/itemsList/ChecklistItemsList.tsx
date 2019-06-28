import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { WithChecklistQueryResults } from '../../components/withChecklistQueryResults/WithChecklistQueryResults'
import { ChecklistAreaContext } from '../ChecklistArea'
import { ChecklistListItem } from '../item/ChecklistListItem'

export interface ChecklistItemsListContext {
    itemClassName?: string
}

interface Props
    extends Partial<QueryParameterProps>,
        ChecklistItemsListContext,
        ChecklistAreaContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The list of checklist items.
 */
export const ChecklistItemsList: React.FunctionComponent<Props> = ({ itemClassName, query, ...props }) => (
    <WithChecklistQueryResults {...props} query={query}>
        {({ checklistOrError }) => (
            <div className="checklist-items-list">
                {isErrorLike(checklistOrError) ? (
                    <div className={itemClassName}>
                        <div className="alert alert-danger mt-2">{checklistOrError.message}</div>
                    </div>
                ) : checklistOrError === LOADING ? (
                    <div className={itemClassName}>
                        <LoadingSpinner className="mt-3" />
                    </div>
                ) : checklistOrError.length === 0 ? (
                    <div className={itemClassName}>
                        <p className="p-2 mb-0 text-muted">No checklists found.</p>
                    </div>
                ) : (
                    <ul className="list-group list-group-flush mb-0">
                        {checklistOrError.map((checklist, i) => (
                            <li key={i} className="list-group-item px-0">
                                <ChecklistListItem
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
    </WithChecklistQueryResults>
)

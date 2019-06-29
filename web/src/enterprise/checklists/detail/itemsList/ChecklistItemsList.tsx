import H from 'history'
import React from 'react'
import { ChecklistAreaContext } from '../ChecklistArea'
import { ChecklistListItem } from '../item/ChecklistListItem'

export interface ChecklistItemsListContext {
    itemClassName?: string
}

interface Props extends ChecklistItemsListContext, ChecklistAreaContext {
    history: H.History
    location: H.Location
}

/**
 * The list of checklist items.
 */
export const ChecklistItemsList: React.FunctionComponent<Props> = ({ itemClassName, checklist, ...props }) => (
    <div className="checklist-items-list">
        <ul className="list-group list-group-flush mb-0">
            {checklist.items.map((item, i) => (
                <li key={i} className="list-group-item px-0">
                    <ChecklistListItem
                        {...props}
                        key={JSON.stringify(item)}
                        item={item}
                        className={itemClassName}
                        headerClassName="pl-5"
                    />
                </li>
            ))}
        </ul>
    </div>
)

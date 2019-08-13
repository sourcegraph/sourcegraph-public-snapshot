import H from 'history'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { ThreadListContext } from './ThreadList'

const LOADING = 'loading' as const

export interface ThreadListHeaderContext extends Pick<QueryParameterProps, 'query'> {
    threads: typeof LOADING | GQL.IThreadConnection | GQL.IThreadOrThreadPreviewConnection | ErrorLike
    location: H.Location
}

export interface ThreadListHeaderItems {
    /**
     * A React fragment to render on the left side of the list header.
     */
    left?: React.ReactFragment

    /**
     * A React fragment to render on the right side of the list header.
     */
    right?: React.ReactFragment
}

interface Props extends ThreadListContext, ThreadListHeaderContext {
    items?: ThreadListHeaderItems
}

/**
 * The header for the thread list.
 */
export const ThreadListHeader: React.FunctionComponent<Props> = ({ items, itemCheckboxes, ...props }) => (
    <div className="card-header d-flex align-items-center justify-content-between font-weight-normal">
        {itemCheckboxes && (
            <input
                type="checkbox"
                className="form-check mx-1 my-2"
                style={{ verticalAlign: 'middle' }}
                aria-label="Select item"
            />
        )}
        {items && items.left}
        <div className="flex-1" />
        {items && items.right}
    </div>
)

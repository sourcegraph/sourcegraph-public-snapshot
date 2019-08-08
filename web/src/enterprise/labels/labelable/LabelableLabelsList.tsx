import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { Label } from '../../../components/Label'
import { useLabelableLabels } from './useLabelableLabels'

interface Props {
    /** The labelable object whose labels to list. */
    labelable: Pick<GQL.Labelable, '__typename' | 'id'>

    itemClassName?: string
}

const LOADING = 'loading' as const

/**
 * A list of labels that are applied to a labelable object.
 */
export const LabelableLabelsList: React.FunctionComponent<Props> = ({ labelable, itemClassName }) => {
    const [labels] = useLabelableLabels(labelable)
    return labels === LOADING ? (
        <LoadingSpinner className="icon-inline" />
    ) : isErrorLike(labels) ? (
        <div className="alert alert-danger">{labels.message}</div>
    ) : labels.totalCount > 0 ? (
        <ul className="list-inline">
            {labels.nodes.map(label => (
                <Label
                    key={label.id}
                    label={label}
                    tag="li"
                    className={`list-inline-item text-truncate ${itemClassName}`}
                />
            ))}
        </ul>
    ) : (
        <small className="text-muted">No labels</small>
    )
}

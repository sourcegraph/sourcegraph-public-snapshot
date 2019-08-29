import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Label } from '../../../components/Label'

interface Props {
    /** The labelable object. */
    labelable: Pick<GQL.Labelable, 'labels'>

    showNoLabels?: boolean

    className?: string
    itemClassName?: string
}

/**
 * A list of labels that are applied to a labelable object.
 */
export const LabelableLabelsList: React.FunctionComponent<Props> = ({
    labelable: { labels },
    showNoLabels = true,
    className = '',
    itemClassName = '',
}) =>
    labels.totalCount > 0 ? (
        <ul className={`list-inline ${className}`}>
            {labels.nodes.map(label => (
                <Label
                    key={label.name}
                    label={label}
                    tag="li"
                    className={`list-inline-item text-truncate ${itemClassName}`}
                />
            ))}
        </ul>
    ) : showNoLabels ? (
        <small className="text-muted">No labels</small>
    ) : null

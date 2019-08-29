import React from 'react'

export interface ThreadListItemContext {}

interface Props {
    /**
     * Whether each item should have a checkbox.
     */
    itemCheckboxes?: boolean

    left?: React.ReactFragment
    title: React.ReactFragment
    afterTitle?: React.ReactFragment
    detail?: React.ReactFragment[]
    right?: React.ReactFragment[]

    className?: string
}

/**
 * An abstract item in the list of threads. Use {@link ThreadListItem} to render a list item for a
 * concrete thread. This component is used to render things that have the same structure as a thread
 * list item.
 */
export const AbstractThreadListItem: React.FunctionComponent<Props> = ({
    left,
    title,
    afterTitle,
    detail,
    right,
    itemCheckboxes,
    className = '',
}) => (
    <li className={`list-group-item ${className}`}>
        <div className="d-flex align-items-start">
            {itemCheckboxes && (
                <div
                    className="form-check ml-1 mr-2"
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ marginTop: '4px' }}
                >
                    <input className="form-check-input position-static" type="checkbox" aria-label="Select item" />
                </div>
            )}
            {left && <div className="mt-1 mr-2">{left}</div>}
            <div className="flex-1">
                <div className="d-flex align-items-center flex-wrap">
                    <h3 className="d-flex align-items-center mb-0 mr-2">{title}</h3>
                    {afterTitle}
                </div>
                {detail && detail.length > 0 && (
                    <ul className="list-inline d-flex align-items-center small text-muted mb-0">
                        {detail.map((e, i) => (
                            // eslint-disable-next-line react/no-array-index-key
                            <li key={i} className="list-inline-item">
                                {e}
                            </li>
                        ))}
                    </ul>
                )}
            </div>
            {right && right.length > 0 && (
                <div>
                    <ul className="list-inline d-flex align-items-center">
                        {right.map((e, i) => (
                            // eslint-disable-next-line react/no-array-index-key
                            <li key={i} className="list-inline-item">
                                {e}
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    </li>
)

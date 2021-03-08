import * as React from 'react'
import classnames from 'classnames'
import { AggregableTag } from 'sourcegraph'

export interface AggregatedTagProps {
    tag: AggregableTag
    className?: string
}

export const AggregatedTag: React.FunctionComponent<AggregatedTagProps> = ({ tag: { text, linkURL }, className }) => {
    const tagComponent = <span className={classnames('badge badge-secondary aggregated-tag', className)}>{text}</span>

    return linkURL ? (
        <a href="{linkURL}" target="_blank">
            {tagComponent}
        </a>
    ) : (
        tagComponent
    )
}

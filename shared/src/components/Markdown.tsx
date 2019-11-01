import classNames from 'classnames'
import * as React from 'react'

interface Props {
    dangerousInnerHTML: string
    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (ref: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = (props: Props) => (
    <div
        ref={props.refFn}
        className={classNames(props.className, 'markdown')}
        dangerouslySetInnerHTML={{ __html: props.dangerousInnerHTML }}
    />
)

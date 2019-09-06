import * as React from 'react'

interface Props {
    dangerousInnerHTML: string

    /**
     * Treat the Markdown as an inline fragment, with no paragraph breaks or margins.
     */
    inline?: boolean

    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (ref: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = (props: Props) => {
    const Tag: 'span' | 'div' = props.inline ? 'span' : 'div'
    return (
        <Tag
            ref={props.refFn}
            className={`markdown ${props.inline ? 'markdown--inline' : ''} ${props.className}`}
            dangerouslySetInnerHTML={{ __html: props.dangerousInnerHTML }}
        />
    )
}

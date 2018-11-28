import * as React from 'react'

interface Props {
    dangerousInnerHTML: string
    className?: string
    refFn?: (ref: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = (props: Props) => (
    <div
        ref={props.refFn}
        className={`markdown ${props.className}`}
        dangerouslySetInnerHTML={{ __html: props.dangerousInnerHTML }}
    />
)

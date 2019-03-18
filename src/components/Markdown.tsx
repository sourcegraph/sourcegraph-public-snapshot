import * as React from 'react'

interface Props {
    dangerousInnerHTML: string
}

export const Markdown: React.StatelessComponent<Props> = (props: Props) => (
    <div className="markdown" dangerouslySetInnerHTML={{ __html: props.dangerousInnerHTML }} />
)

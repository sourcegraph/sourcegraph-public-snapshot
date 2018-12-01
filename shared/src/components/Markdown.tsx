import * as React from 'react'

interface Props {
    dangerousInnerHTML: string
}

export const Markdown: React.FunctionComponent<Props> = (props: Props) => (
    <div className="markdown" dangerouslySetInnerHTML={{ __html: props.dangerousInnerHTML }} />
)

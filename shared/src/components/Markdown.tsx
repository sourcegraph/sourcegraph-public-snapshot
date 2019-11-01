import classNames from 'classnames'
import * as React from 'react'
import sanitizeHtml from 'sanitize-html'

interface Props {
    dangerousInnerHTML: string
    className?: string
    /** Used to strip off any HTML, useful for previews where no formatting is allowed */
    plainText?: boolean
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (ref: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = (props: Props) => (
    <div
        ref={props.refFn}
        className={classNames(props.className, 'markdown')}
        dangerouslySetInnerHTML={{
            __html: props.plainText
                ? sanitizeHtml(props.dangerousInnerHTML, { allowedTags: [], allowedAttributes: {} })
                : props.dangerousInnerHTML,
        }}
    />
)

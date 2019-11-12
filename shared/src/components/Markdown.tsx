import classNames from 'classnames'
import * as React from 'react'

interface Props {
    wrapper?: 'div' | 'span'
    dangerousInnerHTML: string
    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (ref: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = ({
    wrapper: RootComponent = 'div',
    refFn,
    className,
    dangerousInnerHTML,
}: Props) => (
    <RootComponent
        ref={refFn}
        className={classNames(className, 'markdown')}
        dangerouslySetInnerHTML={{ __html: dangerousInnerHTML }}
    />
)

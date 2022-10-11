import React, { useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import * as uuid from 'uuid'

import { NotebookComponent, NotebookComponentProps } from '../../notebooks/notebook/NotebookComponent'
import { convertMarkdownToBlocks } from '../../notebooks/serialize/convertMarkdownToBlocks'

import styles from './RenderedNotebookMarkdown.module.scss'

interface RenderedNotebookMarkdownProps extends Omit<NotebookComponentProps, 'onSerializeBlocks' | 'blocks'> {
    markdown: string
    className?: string
}

export const RenderedNotebookMarkdown: React.FunctionComponent<
    React.PropsWithChildren<RenderedNotebookMarkdownProps>
> = ({ markdown, className, ...props }) => {
    // Generate fresh block IDs, since we do not store them in Markdown.
    const blocks = useMemo(() => convertMarkdownToBlocks(markdown).map(block => ({ id: uuid.v4(), ...block })), [
        markdown,
    ])
    return (
        <div className={classNames(styles.renderedSearchNotebookMarkdownWrapper, className)}>
            <div className={styles.renderedSearchNotebookMarkdown}>
                <NotebookComponent isReadOnly={true} blocks={blocks} {...props} onSerializeBlocks={noop} />
            </div>
        </div>
    )
}

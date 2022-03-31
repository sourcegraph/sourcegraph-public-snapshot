import React, { useState, useMemo } from 'react'

import classNames from 'classnames'
// We're using marked import here to access the type definition, not to render markdown.
// eslint-disable-next-line no-restricted-imports
import { marked } from 'marked'
import ReactDOM from 'react-dom'

import { markdownLexer } from '@sourcegraph/common'

import { Block, MarkdownBlock } from '..'

import styles from './NotebookOutline.module.scss'

interface NotebookOutlineProps {
    notebookElement: HTMLDivElement | null
    blocks: Block[]
}

const OUTLINE_MIN_WIDTH = 200
const OUTLINE_MAX_WIDTH = 400

// TODO: Hide on Blob page
export const NotebookOutline: React.FunctionComponent<NotebookOutlineProps> = React.memo(
    ({ notebookElement, blocks }) => {
        const [availableWidth, setAvailableWidth] = useState(0)

        // const measureAvailableWidth = useCallback(() => {
        //     if (!notebookElement) {
        //         return
        //     }
        //     const notebookRectangle = notebookElement.getBoundingClientRect()
        //     setAvailableWidth(notebookRectangle.left)
        // }, [notebookElement, setAvailableWidth])

        // useEffect(() => measureAvailableWidth(), [measureAvailableWidth])

        // useEffect(() => {
        //     const onResize = (): void => {
        //         window.requestAnimationFrame(measureAvailableWidth)
        //     }
        //     window.addEventListener('resize', onResize)
        //     return () => window.removeEventListener('resize', onResize)
        // }, [measureAvailableWidth])

        const headings = useMemo(
            () =>
                blocks
                    .filter((block): block is MarkdownBlock => block.type === 'md')
                    .flatMap(block => markdownLexer(block.input.text))
                    .filter((token): token is marked.Tokens.Heading => token.type === 'heading' && token.depth <= 2),
            [blocks]
        )

        // if (availableWidth < OUTLINE_MIN_WIDTH) {
        //     return null
        // }
        const outlineContainer = document.querySelector<HTMLDivElement>('[data-notebook-outline-container]')
        if (!outlineContainer) {
            return null
        }

        return ReactDOM.createPortal(
            <div
                className={styles.outline}
                // eslint-disable-next-line react/forbid-dom-props
                // style={{ /* left: `-${availableWidth}px`,*/ width: `${Math.min(OUTLINE_MAX_WIDTH, availableWidth)}px` }}
            >
                {headings.map((heading, index) => (
                    <div
                        // We have to use the index in case of duplicate headings
                        // eslint-disable-next-line react/no-array-index-key
                        key={`${heading.text}_${index}`}
                        className={classNames(styles.heading, `heading-${heading.depth}`)}
                    >
                        {heading.text}
                    </div>
                ))}
            </div>,
            outlineContainer
        )
    }
)

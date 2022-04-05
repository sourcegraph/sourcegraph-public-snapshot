import React, { useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
// We're using marked import here to access the `marked` package type definitions.
// eslint-disable-next-line no-restricted-imports
import { marked, Slugger } from 'marked'
import ReactDOM from 'react-dom'

import { markdownLexer } from '@sourcegraph/common'
import { Link } from '@sourcegraph/wildcard'

import { Block, MarkdownBlock } from '..'

import styles from './NotebookOutline.module.scss'

interface NotebookOutlineProps {
    notebookElement: HTMLElement
    outlineContainerElement: HTMLElement
    blocks: Block[]
}

export const NotebookOutline: React.FunctionComponent<NotebookOutlineProps> = React.memo(
    ({ notebookElement, outlineContainerElement, blocks }) => {
        const scrollableContainer = useRef<HTMLDivElement>(null)
        const [visibleHeadings, setVisibleHeadings] = useState<string[]>([])

        const headings = useMemo(
            () =>
                blocks
                    .filter((block): block is MarkdownBlock => block.type === 'md')
                    .flatMap(block => {
                        const slugger = new Slugger()
                        return markdownLexer(block.input.text)
                            .filter(
                                (token): token is marked.Tokens.Heading => token.type === 'heading' && token.depth <= 2
                            )
                            .map(token => ({
                                text: token.text,
                                id: `${slugger.slug(token.text)}-${block.id}`,
                                depth: token.depth,
                            }))
                    }),
            [blocks]
        )

        useEffect(() => {
            const observeHeadings = (): void => {
                for (const element of notebookElement.querySelectorAll('h1, h2')) {
                    intersectionObserver.observe(element)
                }
            }

            const scrollToOutlineHeading = (id: string): void =>
                document
                    .querySelector(`[data-id="${id}"]`)
                    ?.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'start' })

            const visibleHeadingsSet = new Set<string>()

            const processIntersectionEntries = (entries: IntersectionObserverEntry[]): void => {
                let hasScrolled = false
                // TODO: Improve
                for (const entry of entries) {
                    const headingId = entry.target.id
                    if (entry.isIntersecting) {
                        visibleHeadingsSet.add(headingId)
                        if (!hasScrolled) {
                            // scrollToOutlineHeading(headingId)
                            hasScrolled = true
                        }
                    } else {
                        visibleHeadingsSet.delete(headingId)
                    }
                }
                setVisibleHeadings([...visibleHeadingsSet])
            }

            const intersectionCallback = (entries: IntersectionObserverEntry[]): number =>
                window.requestAnimationFrame(() => processIntersectionEntries(entries))

            const intersectionObserver = new IntersectionObserver(intersectionCallback)
            observeHeadings()

            const mutationObserver = new MutationObserver(observeHeadings)
            mutationObserver.observe(notebookElement, { childList: true, subtree: true })

            return () => {
                intersectionObserver.disconnect()
                mutationObserver.disconnect()
            }
        }, [notebookElement, scrollableContainer, setVisibleHeadings])

        return ReactDOM.createPortal(
            <div className={styles.outline}>
                <div className={styles.title}>Outline</div>
                <div className={styles.scrollableContainer} ref={scrollableContainer}>
                    {headings.map(heading => (
                        <div
                            key={heading.id}
                            data-id={heading.id}
                            className={classNames(
                                styles.heading,
                                `heading-${heading.depth}`,
                                visibleHeadings.includes(heading.id) && styles.visible
                            )}
                        >
                            <Link className={classNames(styles.headingLink)} to={`#${heading.id}`}>
                                {heading.text}
                            </Link>
                        </div>
                    ))}
                </div>
            </div>,
            outlineContainerElement
        )
    }
)

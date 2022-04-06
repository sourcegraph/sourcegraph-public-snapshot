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

        const headingsOrder = useMemo(
            () =>
                headings.reduce((accumulator, current, index) => {
                    accumulator.set(current.id, index)
                    return accumulator
                }, new Map<string, number>()),

            [headings]
        )

        const [highlightedHeading, setHighlightedHeading] = useState<string>('')
        useEffect(() => {
            if (visibleHeadings.length === 0) {
                // If there are no visible headings, keep the last highlighted heading.
                return
            }
            const firstVisibleHeading = visibleHeadings.reduce((accumulator, current) => {
                const currentIndex = headingsOrder.get(current) ?? -1
                const accumulatorIndex = headingsOrder.get(accumulator) ?? -1
                return currentIndex < accumulatorIndex ? current : accumulator
            }, visibleHeadings[0])
            setHighlightedHeading(firstVisibleHeading)
        }, [headingsOrder, visibleHeadings, setHighlightedHeading])

        useEffect(() => {
            const observeHeadings = (): void => {
                for (const element of notebookElement.querySelectorAll('h1, h2')) {
                    intersectionObserver.observe(element)
                }
            }

            // Keep a local copy of the visible headings as a set. This allows us to update
            // the component state in a single `setVisibleHeadings` call.
            const visibleHeadingsSet = new Set<string>()
            const processIntersectionEntries = (entries: IntersectionObserverEntry[]): void => {
                for (const entry of entries) {
                    const headingId = entry.target.id
                    if (entry.isIntersecting) {
                        visibleHeadingsSet.add(headingId)
                    } else {
                        visibleHeadingsSet.delete(headingId)
                    }
                }
                setVisibleHeadings([...visibleHeadingsSet])
            }

            // We use the requestAnimationFrame callback to avoid processing on the main thread.
            const intersectionCallback = (entries: IntersectionObserverEntry[]): number =>
                window.requestAnimationFrame(() => processIntersectionEntries(entries))

            // We use the IntersectionObserver to keep track of the headings that
            // are currently in the viewport.
            const intersectionObserver = new IntersectionObserver(intersectionCallback)
            observeHeadings()

            // On every notebook mutation, observe the rendered headings again.
            const mutationObserver = new MutationObserver(observeHeadings)
            mutationObserver.observe(notebookElement, { childList: true, subtree: true })

            return () => {
                intersectionObserver.disconnect()
                mutationObserver.disconnect()
            }
        }, [notebookElement, setVisibleHeadings])

        useEffect(() => {
            if (!highlightedHeading || !scrollableContainer.current) {
                return
            }
            const heading = scrollableContainer.current.querySelector<HTMLElement>(`[data-id="${highlightedHeading}"]`)
            if (!heading) {
                return
            }
            // Scroll to the highlighted heading in the outline.
            const top = heading.offsetTop > scrollableContainer.current.offsetHeight ? heading.offsetTop : 0
            scrollableContainer.current.scrollTo({ top, behavior: 'smooth' })
        }, [highlightedHeading])

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
                                highlightedHeading === heading.id && styles.highlight
                            )}
                            data-test-highlighted={highlightedHeading === heading.id}
                        >
                            <Link className={classNames(styles.headingLink)} to={`#${heading.id}`}>
                                {highlightedHeading === heading.id && (
                                    <span className={styles.highlightDot}>&middot;</span>
                                )}
                                <span>{heading.text}</span>
                            </Link>
                        </div>
                    ))}
                </div>
            </div>,
            outlineContainerElement
        )
    }
)

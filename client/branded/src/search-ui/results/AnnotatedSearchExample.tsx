import React from 'react'

import { mdiFormatLetterCase, mdiRegex, mdiCodeBrackets, mdiMagnify } from '@mdi/js'

import { Icon } from '@sourcegraph/wildcard'

import styles from './AnnotatedSearchExample.module.scss'

const arrowHeight = 27
const edgesHeight = 8
const aboveArrowY = 70
const belowArrowY = 150

/**
 * Helper function to create "arrows" of the form
 *
 *         |                     |__________|
 *   ______|______                    |
 *  |            |                    |
 *
 * The function takes the X coordinate, the width and the position relative to
 * the input (above/below) as argument and computes the position and dimensions
 * of the four <line>s that make up this "arrow".
 */

function arrow(x: number, width: number, position: 'above' | 'below'): React.ReactElement {
    const pointerLine = <line x1={width / 2} x2={width / 2} y1="0" y2={arrowHeight} />
    const centerLine = <line x1="0" x2={width} y1={arrowHeight} y2={arrowHeight} />
    const leftEdge = <line x1="0" x2="0" y1={arrowHeight} y2={arrowHeight + edgesHeight} />
    const rightEdge = <line x1={width} x2={width} y1={arrowHeight} y2={arrowHeight + edgesHeight} />

    let group = (
        <>
            {pointerLine}
            {centerLine}
            {leftEdge}
            {rightEdge}
        </>
    )

    // To keep things simple we use the same composition of lines of above and
    // below arrows and simple rotate the arrows below the search input by 180
    // degrees.
    if (position === 'below') {
        group = <g transform={`rotate(180, ${width / 2}, ${(arrowHeight + edgesHeight) / 2})`}>{group}</g>
    }
    return (
        <g transform={`translate(${x}, ${position === 'above' ? aboveArrowY : belowArrowY})`} className={styles.arrow}>
            {' '}
            {group}
        </g>
    )
}

export const AnnotatedSearchInput: React.FunctionComponent<React.PropsWithChildren<{ showSearchContext: boolean }>> =
    React.memo(({ showSearchContext }) => {
        // I'd like to say that there is a logic behind these numbers but it's
        // mostly trial and error.
        const inputStart = showSearchContext ? 56.5 : 178
        const filterTextStart = 116 + (showSearchContext ? 0 : 30)
        const viewBoxX = showSearchContext ? 55 : filterTextStart
        const width = 736 - viewBoxX
        const height = showSearchContext ? 222 : 222

        return (
            // the viewBox is adjusted to "crop" the image to its content
            // Original width and height of the image was 800x270
            <svg
                className={styles.annotatedSearchInput}
                width={width}
                height={height}
                viewBox={`${viewBoxX} 35 ${width} ${height}`}
                preserveAspectRatio="xMidYMid meet"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
            >
                <g>
                    <path
                        d={`M ${inputStart} 113 c 0 -1.381 1.1193 -2.5 2.5 -2.5 H 688.5 v 33 H ${
                            inputStart + 2.5
                        } c -1.3807 0 -2.5 -1.119 -2.5 -2.5 v -28 z`}
                        className={styles.searchBox}
                    />
                    {showSearchContext && (
                        <>
                            <text className={styles.code} x="68" y="130.836">
                                <tspan className={styles.filter}>context:</tspan>
                                <tspan>global</tspan>
                            </text>
                            <line className={styles.separator} x1="178" y1="118" x2="178" y2="137" />
                        </>
                    )}
                    <text className={styles.code} x="171.852" y="130.836">
                        <tspan className={styles.filter}>{'  '}repo:</tspan>
                        <tspan>sourcegraph/sourcegraph</tspan>
                        <tspan className={styles.metaRegexpCharacterSet}>.</tspan>
                        <tspan className={styles.metaRegexpRangeQuantifier}>*</tspan>
                        <tspan> function auth(){'{'} </tspan>
                    </text>
                    <Icon aria-label="Case sensitivity toggle" x="590" y="115" svgPath={mdiFormatLetterCase} />
                    <Icon aria-label="Regular expression toggle" x="620" y="115" svgPath={mdiRegex} />
                    <Icon aria-label="Structural search toggle" x="650" y="115" svgPath={mdiCodeBrackets} />
                    <path
                        d="M688 110H731C732.105 110 733 110.895 733 112V142C733 143.105 732.105 144 731 144H688V110Z"
                        fill="#1475CF"
                    />
                    <Icon aria-label="Search" className={styles.searchIcon} x="698" y="115" svgPath={mdiMagnify} />

                    {arrow(188, 30, 'above')}
                    <text transform={`translate(${filterTextStart}, 44)`}>
                        <tspan x="0" y="0">
                            Filters scope your search to repos,{' '}
                        </tspan>
                        <tspan x="0" y="16">
                            orgs, languages, and more.
                        </tspan>
                    </text>

                    {arrow(410, 120, 'above')}
                    <text transform="translate(395, 44)">
                        <tspan x="0" y="0">
                            By default, search terms are{' '}
                        </tspan>
                        <tspan x="0" y="16">
                            interpreted literally (without regexp).
                        </tspan>
                    </text>

                    {showSearchContext && (
                        <>
                            {arrow(68, 108, 'below')}
                            <text transform="translate(56, 200)">
                                <tspan x="0" y="0">
                                    By default, Sourcegraph searches the{' '}
                                </tspan>
                                <tspan x="0" y="16">
                                    <tspan className={styles.bold}>global </tspan>
                                    context, which is publicly
                                </tspan>
                                <tspan x="0" y="32">
                                    available code on code hosts such as{' '}
                                </tspan>
                                <tspan x="0" y="48">
                                    GitHub and GitLab.
                                </tspan>
                            </text>
                        </>
                    )}

                    {arrow(387, 16, 'below')}
                    <text transform="translate(340, 200)">
                        <tspan x="0" y="0">
                            You can use regexp inside{' '}
                        </tspan>
                        <tspan x="0" y="16">
                            filters, even when not in{' '}
                        </tspan>
                        <tspan x="0" y="32">
                            regexp mode
                        </tspan>
                    </text>

                    {arrow(590, 82, 'below')}
                    <text transform="translate(538, 200)">
                        <tspan x="0" y="0">
                            Search can be case-sensitive{' '}
                        </tspan>
                        <tspan x="0" y="16">
                            one of three modes: literal{' '}
                        </tspan>
                        <tspan x="0" y="32">
                            (default), regexp or structural.
                        </tspan>
                    </text>
                </g>
            </svg>
        )
    })

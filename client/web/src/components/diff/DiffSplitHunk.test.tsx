import { cleanup, fireEvent, render, type RenderResult } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import { DiffHunkLineType, type FileDiffHunkFields } from '../../graphql-operations'

import { type DiffHunkProps, DiffSplitHunk } from './DiffSplitHunk'

import lineStyles from './Lines.module.scss'

describe('DiffSplitHunk', () => {
    const hunk: FileDiffHunkFields = {
        oldRange: { startLine: 159, lines: 7 },
        oldNoNewlineAt: false,
        newRange: { startLine: 159, lines: 7 },
        section: 'export async function register({',
        highlight: {
            aborted: false,
            lines: [
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        const subscriptions = new Subscription()',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        const foo = sourcegraph.app.foo()',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        const connection = await createConnection()',
                },
                {
                    kind: DiffHunkLineType.DELETED,
                    html: '        logger.log(`WebSocket connection to TypeScript backend opened`)',
                },
                {
                    kind: DiffHunkLineType.ADDED,
                    html: '        logger.log(`WebSocket connection to language server opened`)',
                },
                {
                    kind: DiffHunkLineType.ADDED,
                    html: '        subscriptions.add(',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '                connection.observeNotification(LogMessageNotification.type).subscribe(({ type, message }) =\u003E {',
                },
                {
                    kind: DiffHunkLineType.DELETED,
                    html: '                    const method = LSP_TO_LOG_LEVEL[type]',
                },
            ],
        },
    }

    let queries: RenderResult
    const renderWithProps = (props: DiffHunkProps): RenderResult =>
        render(
            <MemoryRouter>
                <table>
                    <tbody>
                        <DiffSplitHunk {...props} />
                    </tbody>
                </table>
            </MemoryRouter>
        )

    afterEach(cleanup)

    describe('Split Lines Diff', () => {
        beforeEach(() => {
            queries = renderWithProps({
                hunk,
                lineNumbers: true,
                fileDiffAnchor: 'anchor_',
            })
        })
        it('will show a DELETED on the left and the ADDED on the right', () => {
            const diffLine = queries.getByTestId('anchor_L162')
            expect(diffLine).toBeInTheDocument()

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[0]).toHaveTextContent('162')
            expect(diffLine.children[1]).toHaveTextContent(
                'logger.log(`WebSocket connection to TypeScript backend opened`)'
            )
            expect(diffLine.children[2]).toHaveTextContent('162')
            expect(diffLine.children[3]).toHaveTextContent(
                'logger.log(`WebSocket connection to language server opened`)'
            )
        })

        it('will show a single ADDED on the right and empty cell on the left', () => {
            const diffLine = queries.getByTestId('anchor_R163')
            expect(diffLine).toBeInTheDocument()

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[0]).toHaveTextContent('')
            expect(diffLine.children[1]).toHaveTextContent('')
            expect(diffLine.children[2]).toHaveTextContent('163')
            expect(diffLine.children[3]).toHaveTextContent('subscriptions.add(')
        })

        it('will show a single DELETED on the left and empty cell on the right', () => {
            const diffLine = queries.getByTestId('anchor_L164')
            expect(diffLine).toBeInTheDocument()

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[0]).toHaveTextContent('164')
            expect(diffLine.children[1]).toHaveTextContent('const method = LSP_TO_LOG_LEVEL[type]')
            expect(diffLine.children[2]).toHaveTextContent('')
            expect(diffLine.children[3]).toHaveTextContent('')
        })

        it('will show UNCHANGED lines on both sides', () => {
            const diffLine = queries.getByTestId('anchor_L159')
            expect(diffLine).toBeInTheDocument()

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[0]).toHaveTextContent('159')
            expect(diffLine.children[1]).toHaveTextContent('const subscriptions = new Subscription()')
            expect(diffLine.children[2]).toHaveTextContent('159')
            expect(diffLine.children[1]).toHaveTextContent('const subscriptions = new Subscription()')
        })

        it('add active class when the anchor is clicked on UNCHANGED lines', () => {
            const diffLine = queries.getByTestId('anchor_L159')
            fireEvent.click(diffLine.children[0].children[0])

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[1]).toHaveClass(lineStyles.lineActive)
            expect(diffLine.children[3]).toHaveClass(lineStyles.lineActive)
        })

        it('add active class when the anchor is clicked on ADDED lines', () => {
            const diffLine = queries.getByTestId('anchor_R163')
            fireEvent.click(diffLine.children[2].children[0])

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[1]).not.toHaveClass(lineStyles.lineActive)
            expect(diffLine.children[3]).toHaveClass(lineStyles.lineActive)
        })

        it('add active class when the anchor is clicked on DELETED lines', () => {
            const diffLine = queries.getByTestId('anchor_L164')
            fireEvent.click(diffLine.children[0].children[0])

            expect(diffLine.children).toHaveLength(4)
            expect(diffLine.children[1]).toHaveClass(lineStyles.lineActive)
            expect(diffLine.children[3]).not.toHaveClass(lineStyles.lineActive)
        })
    })
})

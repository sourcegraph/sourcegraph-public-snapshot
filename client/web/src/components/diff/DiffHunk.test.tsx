import { render } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router-dom'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { DiffHunkLineType, FileDiffHunkFields } from '../../graphql-operations'

import { DiffHunk } from './DiffHunk'

describe('DiffHunk', () => {
    const history = createMemoryHistory()
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
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '        subscriptions.add(',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '                connection.observeNotification(LogMessageNotification.type).subscribe(({ type, message }) =\u003E {',
                },
                {
                    kind: DiffHunkLineType.UNCHANGED,
                    html: '                    const method = LSP_TO_LOG_LEVEL[type]',
                },
            ],
        },
    }

    it('renders a unified diff view for the given diff hunk', () => {
        expect(
            render(
                <Router history={history}>
                    <CompatRouter>
                        <table>
                            <tbody>
                                <DiffHunk hunk={hunk} lineNumbers={true} fileDiffAnchor="anchor_" />
                            </tbody>
                        </table>
                    </CompatRouter>
                </Router>
            ).asFragment()
        ).toMatchSnapshot()
    })
})

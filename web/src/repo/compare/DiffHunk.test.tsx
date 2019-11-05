import { Range } from '@sourcegraph/extension-api-classes'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import * as GQL from '../../../../shared/src/graphql/schema'
import { DiffHunk } from './DiffHunk'

describe('DiffHunk', () => {
    const history = H.createMemoryHistory()

    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    const hunk = {
        oldRange: { startLine: 159, lines: 7 },
        oldNoNewlineAt: false,
        newRange: { startLine: 159, lines: 7 },
        section: 'export async function register({',
        body:
            '         const subscriptions = new Subscription()\n         const decorationType = sourcegraph.app.createDecorationType()\n         const connection = await createConnection()\n-        logger.log(`WebSocket connection to TypeScript backend opened`)\n+        logger.log(`WebSocket connection to language server opened`)\n         subscriptions.add(\n             connection.observeNotification(LogMessageNotification.type).subscribe(({ type, message }) =\u003e {\n                 const method = LSP_TO_LOG_LEVEL[type]\n',
    } as GQL.IFileDiffHunk

    it('renders a unified diff view for the given diff hunk', () => {
        expect(
            renderer
                .create(
                    <DiffHunk
                        hunk={hunk}
                        decorations={{ head: new Map(), base: new Map() }}
                        lineNumbers={true}
                        isLightTheme={true}
                        fileDiffAnchor="anchor_"
                        history={history}
                        location={H.createLocation('/testdiff', history.location)}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    const decorations = {
        head: new Map<number, TextDocumentDecoration[]>([
            [
                159,
                [
                    {
                        range: new Range(158, 0, 158, 0),
                        isWholeLine: true,
                        backgroundColor: 'red',
                        dark: { border: '1px blue solid' },
                        after: {
                            contentText: 'head content',
                            hoverMessage: 'base hover msg',
                            backgroundColor: 'black',
                        },
                    },
                ],
            ],
        ]),
        base: new Map<number, TextDocumentDecoration[]>([
            [
                162,
                [
                    {
                        range: new Range(161, 0, 161, 0),
                        isWholeLine: true,
                        backgroundColor: 'blue',
                        dark: { border: '1px blue solid' },
                        after: {
                            contentText: 'base content',
                            hoverMessage: 'base hover msg',
                            backgroundColor: 'black',
                        },
                    },
                ],
            ],
        ]),
    }

    it('renders decorations if given', () => {
        expect(
            renderer
                .create(
                    <DiffHunk
                        hunk={hunk}
                        decorations={decorations}
                        lineNumbers={true}
                        isLightTheme={true}
                        fileDiffAnchor="anchor_"
                        history={history}
                        location={H.createLocation('/testdiff', history.location)}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    it('renders dark theme decorations if dark theme is active', () => {
        expect(
            renderer
                .create(
                    <DiffHunk
                        hunk={hunk}
                        decorations={decorations}
                        lineNumbers={true}
                        isLightTheme={false}
                        fileDiffAnchor="anchor_"
                        history={history}
                        location={H.createLocation('/testdiff', history.location)}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})

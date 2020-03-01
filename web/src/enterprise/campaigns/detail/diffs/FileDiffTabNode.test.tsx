import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { FileDiffTabNode } from './FileDiffTabNode'
import * as GQL from '../../../../../shared/src/graphql/schema'

describe('FileDiffTabNode', () => {
    const renderTest = (persistLines: boolean): void => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation('/campaigns/new')
        expect(
            renderer
                .create(
                    <FileDiffTabNode
                        isLightTheme={true}
                        history={history}
                        location={location}
                        persistLines={persistLines}
                        node={
                            {
                                __typename: 'ChangesetPlan' as const,
                                repository: {
                                    url: 'github.com/sourcegraph/sourcegraph',
                                    name: 'sourcegraph/sourcegraph',
                                } as GQL.IRepository,
                                diff: {
                                    __typename: 'PreviewRepositoryComparison' as const,
                                    fileDiffs: {
                                        nodes: [
                                            {
                                                __typename: 'PreviewFileDiff' as const,
                                                oldPath: 'javascripts/highlight.customized/CHANGES.md',
                                                newPath: 'javascripts/highlight.customized/CHANGES.md',
                                                hunks: [
                                                    {
                                                        oldRange: { startLine: 913, lines: 7 },
                                                        oldNoNewlineAt: false,
                                                        newRange: { startLine: 913, lines: 7 },
                                                        section: 'Fixes for existing languages:',
                                                        body:
                                                            ' \n The highlighter has become more usable as a library allowing to do highlighting\n from initialization code of JS frameworks and in ajax methods (see.\n-readme.eng.txt).\n+donotreadme.eng.txt).\n \n Also this version drops support for the [WordPress][wp] plugin. Everyone is\n welcome to [pick up its maintenance][p] if needed.\n',
                                                    },
                                                ],
                                                stat: { added: 0, changed: 1, deleted: 0 },
                                                internalID: '2e9b56fa49c294b9455f07a509cfab33',
                                            },
                                        ] as GQL.IPreviewFileDiff[],
                                        totalCount: 1,
                                        pageInfo: { hasNextPage: false },
                                        diffStat: { added: 0, changed: 1, deleted: 0 },
                                    },
                                } as GQL.IPreviewRepositoryComparison,
                            } as GQL.IChangesetPlan
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    }
    test('renders the diff node with persisting lines', () => {
        renderTest(true)
    })
    test('renders the diff node without persisting lines', () => {
        renderTest(false)
    })
})

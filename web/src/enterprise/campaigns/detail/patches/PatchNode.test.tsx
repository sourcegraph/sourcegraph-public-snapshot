import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { PatchNode } from './PatchNode'
import { IPatch } from '../../../../../../shared/src/graphql/schema'
import { Subject } from 'rxjs'

describe('PatchNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    const renderPatch = ({ enablePublishing, fileDiff }: { enablePublishing: boolean; fileDiff: boolean }): void => {
        const renderer = createRenderer()
        renderer.render(
            <PatchNode
                isLightTheme={true}
                history={history}
                location={location}
                node={
                    {
                        __typename: 'Patch',
                        diff: fileDiff
                            ? {
                                  fileDiffs: {
                                      __typename: 'FileDiffConnection',
                                      diffStat: {
                                          added: 100,
                                          changed: 200,
                                          deleted: 100,
                                      },
                                  },
                              }
                            : null,
                        repository: {
                            __typename: 'Repository',
                            name: 'sourcegraph',
                            url: 'github.com/sourcegraph/sourcegraph',
                        },
                    } as IPatch
                }
                campaignUpdates={new Subject<void>()}
                enablePublishing={enablePublishing}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    }
    test('renders a patch with publishing disabled', () => {
        renderPatch({ enablePublishing: false, fileDiff: true })
    })
    test('renders a patch with publishing enabled', () => {
        renderPatch({ enablePublishing: true, fileDiff: true })
    })
    test('renders a patch without a filediff', () => {
        renderPatch({ enablePublishing: true, fileDiff: false })
    })
})

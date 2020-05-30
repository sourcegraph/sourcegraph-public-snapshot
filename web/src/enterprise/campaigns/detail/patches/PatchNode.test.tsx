import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { PatchNode } from './PatchNode'
import { IPatch, IFileDiffConnection } from '../../../../../../shared/src/graphql/schema'
import { Subject, of } from 'rxjs'

describe('PatchNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    const renderPatch = (enablePublishing: boolean): void => {
        const renderer = createRenderer()
        renderer.render(
            <PatchNode
                isLightTheme={true}
                history={history}
                location={location}
                node={
                    {
                        __typename: 'Patch',
                        diff: {
                            fileDiffs: {
                                __typename: 'FileDiffConnection',
                                diffStat: {
                                    added: 100,
                                    changed: 200,
                                    deleted: 100,
                                },
                            },
                        },
                        repository: {
                            __typename: 'Repository',
                            name: 'sourcegraph',
                            url: 'github.com/sourcegraph/sourcegraph',
                        },
                    } as IPatch
                }
                campaignUpdates={new Subject<void>()}
                enablePublishing={enablePublishing}
                queryPatchFileDiffs={() => of({ __typename: 'FileDiffConnection' } as IFileDiffConnection)}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    }
    test('renders a patch with publishing disabled', () => {
        renderPatch(false)
    })
    test('renders a patch with publishing enabled', () => {
        renderPatch(true)
    })
})

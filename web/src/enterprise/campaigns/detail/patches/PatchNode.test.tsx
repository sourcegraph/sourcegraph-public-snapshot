import * as H from 'history'
import React from 'react'
import { PatchNode } from './PatchNode'
import { IPatch, IRepository, IPreviewRepositoryComparison } from '../../../../../../shared/src/graphql/schema'
import { Subject } from 'rxjs'
import { shallow } from 'enzyme'

describe('PatchNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    const renderPatch = ({
        enablePublishing,
        publishable,
        fileDiff,
    }: {
        enablePublishing: boolean
        publishable: boolean
        fileDiff: boolean
    }): void => {
        expect(
            shallow(
                <PatchNode
                    isLightTheme={true}
                    history={history}
                    location={location}
                    node={
                        {
                            __typename: 'Patch',
                            id: 'something',
                            publishable,
                            publicationEnqueued: false,
                            diff: fileDiff
                                ? ({
                                      fileDiffs: {
                                          __typename: 'FileDiffConnection',
                                          diffStat: {
                                              added: 100,
                                              changed: 200,
                                              deleted: 100,
                                          },
                                      },
                                  } as IPreviewRepositoryComparison)
                                : null,
                            repository: {
                                __typename: 'Repository',
                                name: 'sourcegraph',
                                url: 'github.com/sourcegraph/sourcegraph',
                            } as IRepository,
                        } as IPatch
                    }
                    campaignUpdates={new Subject<void>()}
                    enablePublishing={enablePublishing}
                />
            )
        ).toMatchSnapshot()
    }
    for (const publishable of [false, true]) {
        test(`renders a patch with publishing disabled and publishable: ${
            publishable ? 'enabled' : 'disabled'
        }`, () => {
            renderPatch({ enablePublishing: false, publishable, fileDiff: true })
        })
        test(`renders a patch with publishing enabled and publishable: ${publishable ? 'enabled' : 'disabled'}`, () => {
            renderPatch({ enablePublishing: true, publishable, fileDiff: true })
        })
    }
    test('renders a patch without a filediff', () => {
        renderPatch({ enablePublishing: true, publishable: true, fileDiff: false })
    })
})

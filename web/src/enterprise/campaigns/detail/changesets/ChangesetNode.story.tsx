import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import { ChangesetNode } from './ChangesetNode'
import { createMemoryHistory } from 'history'
import webStyles from '../../../../SourcegraphWebApp.scss'
import {
    ChangesetState,
    ChangesetReviewState,
    ChangesetCheckState,
    IRepository,
    IHiddenExternalChangeset,
} from '../../../../../../shared/src/graphql/schema'
import { Tooltip } from '../../../../components/tooltip/Tooltip'
import { FILE_DIFF_NODES } from '../../../../components/diff/FileDiffNode.story'
import { of } from 'rxjs'
import { MemoryRouter } from 'react-router'

const { add } = storiesOf('web/ChangesetNode', module).addDecorator(story => {
    // TODO find a way to do this globally for all stories and storybook itself.
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <MemoryRouter>
            <style>{webStyles}</style>
            <Tooltip />
            <div className="p-3 container">{story()}</div>
        </MemoryRouter>
    )
})

add('Hidden changeset', () => (
    <>
        {Object.values(ChangesetState).map((state, index) => (
            <ChangesetNode
                key={index}
                viewerCanAdminister={true}
                node={
                    {
                        __typename: 'HiddenExternalChangeset',
                        state,
                        nextSyncAt: null,
                        updatedAt: new Date().toISOString(),
                    } as IHiddenExternalChangeset
                }
                isLightTheme={true}
                location={createMemoryHistory().location}
                history={createMemoryHistory()}
            />
        ))}
    </>
))

add('Visible changeset', () => (
    <>
        {Object.values(ChangesetState).map((state, index) => (
            <div className="mb-3" key={index}>
                <h3 className="mb-0">{state}</h3>
                <ChangesetNode
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    _queryExternalChangesetWithFileDiffs={() =>
                        of({
                            diff: {
                                __typename: 'PreviewRepositoryComparison',
                                fileDiffs: { nodes: FILE_DIFF_NODES },
                            },
                        } as any)
                    }
                    node={{
                        __typename: 'ExternalChangeset',
                        state,
                        nextSyncAt: null,
                        updatedAt: new Date().toISOString(),
                        base: null,
                        body: 'Body',
                        checkState: radios('checkState', ChangesetCheckState, null),
                        diff: [ChangesetState.PENDING, ChangesetState.PUBLISHING, ChangesetState.OPEN].includes(state)
                            ? ({
                                  __typename: 'PreviewRepositoryComparison',
                                  fileDiffs: { nodes: FILE_DIFF_NODES },
                              } as any)
                            : null,
                        diffStat: {
                            __typename: 'DiffStat',
                            added: 10,
                            changed: 5,
                            deleted: 10,
                        },
                        externalID: [ChangesetState.OPEN, ChangesetState.CLOSED, ChangesetState.MERGED].includes(state)
                            ? '1234'
                            : null,
                        reviewState: radios('reviewState', ChangesetReviewState, ChangesetReviewState.PENDING),
                        labels: [ChangesetState.OPEN, ChangesetState.CLOSED, ChangesetState.MERGED].includes(state)
                            ? [
                                  {
                                      __typename: 'ChangesetLabel',
                                      color: '93ba13',
                                      description: 'Some label',
                                      text: 'Important',
                                  },
                              ]
                            : [],
                        externalURL: [ChangesetState.OPEN, ChangesetState.CLOSED, ChangesetState.MERGED].includes(state)
                            ? {
                                  __typename: 'ExternalLink',
                                  serviceType: 'github',
                                  url: 'http://test.test/changeset/1234',
                              }
                            : null,
                        title: 'Pretty awesome changeset',
                        createdAt: new Date().toISOString(),
                        error: state === ChangesetState.ERRORED ? 'Very bad error' : null,
                        head: null,
                        repository: {
                            name: 'github.com/sourcegraph/awesome',
                        } as IRepository,
                    }}
                    location={createMemoryHistory().location}
                    history={createMemoryHistory()}
                />
            </div>
        ))}
    </>
))

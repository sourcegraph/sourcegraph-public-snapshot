import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import { CampaignUpdateDiff } from './CampaignUpdateDiff'
import webStyles from '../../../SourcegraphWebApp.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { MemoryRouter } from 'react-router'
import {
    ChangesetState,
    IRepository,
    IExternalChangeset,
    IChangesetLabel,
    ChangesetReviewState,
    ICampaign,
} from '../../../../../shared/src/graphql/schema'

const { add } = storiesOf('web/CampaignUpdateDiff', module).addDecorator(story => {
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

add('CampaignDelta not empty', () => (
    <CampaignUpdateDiff
        campaignDelta={{
            __typename: 'CampaignDelta',
            descriptionChanged: boolean('descriptionChanged', false),
            titleChanged: boolean('titleChanged', false),
            viewerCanAdminister: boolean('viewerCanAdminister', true),
            campaign: {
                name: 'Old title',
                description: 'Old description',
            } as ICampaign,
            newTitle: boolean('titleChanged', false) ? 'New title' : 'Old title',
            newDescription: boolean('descriptionChanged', false) ? 'New description' : 'Old description',
            changesets: {
                __typename: 'ChangesetUpdateConnection',
                totalCount: 5,
                pageInfo: {
                    __typename: 'PageInfo',
                    endCursor: null,
                    hasNextPage: false,
                },
                edges: [
                    {
                        __typename: 'ChangesetUpdateEdge',
                        dirty: false,
                        bodyChanged: false,
                        newTitle: 'New title',
                        willPublish: false,
                        node: {
                            __typename: 'ExternalChangeset',
                            state: ChangesetState.OPEN,
                            reviewState: ChangesetReviewState.APPROVED,
                            labels: [] as IChangesetLabel[],
                            externalID: '123',
                            externalURL: {
                                serviceType: 'github',
                                __typename: 'ExternalLink',
                                url: 'http://test.test/123',
                            },
                            title: 'Awesome title',
                            repository: {
                                name: 'github.com/sourcegraph/awesome',
                            } as IRepository,
                            nextSyncAt: null,
                            updatedAt: new Date().toISOString(),
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 3,
                                changed: 1,
                                deleted: 5,
                            },
                        } as IExternalChangeset,
                        newBody: 'New body',
                        patchChanged: false,
                        titleChanged: false,
                        willClose: false,
                        diff: null,
                        newBranch: boolean('branchChanged', false) ? 'the-new-branch-name' : 'the-old-branch-name',
                        oldBranch: 'the-old-branch-name',
                        branchChanged: boolean('branchChanged', false),
                    },
                    {
                        __typename: 'ChangesetUpdateEdge',
                        dirty: true,
                        bodyChanged: false,
                        newTitle: 'Changed the title',
                        willPublish: false,
                        node: {
                            __typename: 'ExternalChangeset',
                            state: ChangesetState.OPEN,
                            reviewState: ChangesetReviewState.PENDING,
                            labels: [] as IChangesetLabel[],
                            title: 'Old title',
                            externalID: '456',
                            externalURL: {
                                serviceType: 'github',
                                __typename: 'ExternalLink',
                                url: 'http://test.test/456',
                            },
                            repository: {
                                name: 'github.com/sourcegraph/awesome',
                            } as IRepository,
                            nextSyncAt: null,
                            updatedAt: new Date().toISOString(),
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 3,
                                changed: 1,
                                deleted: 5,
                            },
                        } as IExternalChangeset,
                        newBody: 'New body',
                        patchChanged: false,
                        titleChanged: true,
                        willClose: false,
                        diff: null,
                        newBranch: boolean('branchChanged', false) ? 'the-new-branch-name' : 'the-old-branch-name',
                        oldBranch: 'the-old-branch-name',
                        branchChanged: boolean('branchChanged', false),
                    },
                    {
                        __typename: 'ChangesetUpdateEdge',
                        dirty: true,
                        bodyChanged: true,
                        newTitle: 'Awesome title',
                        willPublish: false,
                        node: {
                            __typename: 'ExternalChangeset',
                            state: ChangesetState.PENDING,
                            reviewState: ChangesetReviewState.PENDING,
                            labels: [] as IChangesetLabel[],
                            title: 'Awesome title',
                            body: 'Changed the body',
                            repository: {
                                name: 'github.com/sourcegraph/awesome',
                            } as IRepository,
                            nextSyncAt: null,
                            updatedAt: new Date().toISOString(),
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 3,
                                changed: 1,
                                deleted: 5,
                            },
                        } as IExternalChangeset,
                        newBody: 'New body',
                        patchChanged: false,
                        titleChanged: false,
                        willClose: false,
                        diff: null,
                        newBranch: boolean('branchChanged', false) ? 'the-new-branch-name' : 'the-old-branch-name',
                        oldBranch: 'the-old-branch-name',
                        branchChanged: boolean('branchChanged', false),
                    },
                    {
                        __typename: 'ChangesetUpdateEdge',
                        dirty: true,
                        bodyChanged: false,
                        newTitle: 'Awesome title',
                        willPublish: true,
                        node: {
                            __typename: 'ExternalChangeset',
                            state: ChangesetState.PENDING,
                            reviewState: ChangesetReviewState.PENDING,
                            labels: [] as IChangesetLabel[],
                            title: 'Awesome title',
                            body: 'Changed the body',
                            repository: {
                                name: 'github.com/sourcegraph/awesome',
                            } as IRepository,
                            nextSyncAt: null,
                            updatedAt: new Date().toISOString(),
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 3,
                                changed: 1,
                                deleted: 5,
                            },
                        } as IExternalChangeset,
                        newBody: 'New body',
                        patchChanged: false,
                        titleChanged: false,
                        willClose: false,
                        diff: null,
                        newBranch: boolean('branchChanged', false) ? 'the-new-branch-name' : 'the-old-branch-name',
                        oldBranch: 'the-old-branch-name',
                        branchChanged: boolean('branchChanged', false),
                    },
                    {
                        __typename: 'ChangesetUpdateEdge',
                        dirty: true,
                        bodyChanged: false,
                        newTitle: 'Awesome title',
                        willPublish: false,
                        node: {
                            __typename: 'ExternalChangeset',
                            state: ChangesetState.PENDING,
                            reviewState: ChangesetReviewState.PENDING,
                            labels: [] as IChangesetLabel[],
                            title: 'Awesome title',
                            body: 'Changed the body',
                            repository: {
                                name: 'github.com/sourcegraph/awesome',
                            } as IRepository,
                            nextSyncAt: null,
                            updatedAt: new Date().toISOString(),
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 3,
                                changed: 1,
                                deleted: 5,
                            },
                        } as IExternalChangeset,
                        newBody: 'New body',
                        patchChanged: false,
                        titleChanged: false,
                        willClose: true,
                        diff: null,
                        newBranch: boolean('branchChanged', false) ? 'the-new-branch-name' : 'the-old-branch-name',
                        oldBranch: 'the-old-branch-name',
                        branchChanged: boolean('branchChanged', false),
                    },
                    {
                        __typename: 'ChangesetUpdateEdge',
                        dirty: true,
                        bodyChanged: false,
                        newTitle: 'Awesome title',
                        willPublish: false,
                        node: {
                            __typename: 'ExternalChangeset',
                            state: ChangesetState.PENDING,
                            reviewState: ChangesetReviewState.PENDING,
                            labels: [] as IChangesetLabel[],
                            title: 'Awesome title',
                            body: 'Changed the body',
                            repository: {
                                name: 'github.com/sourcegraph/awesome',
                            } as IRepository,
                            nextSyncAt: null,
                            updatedAt: new Date().toISOString(),
                            diffStat: {
                                __typename: 'DiffStat',
                                added: 3,
                                changed: 1,
                                deleted: 5,
                            },
                        } as IExternalChangeset,
                        newBody: 'New body',
                        patchChanged: true,
                        titleChanged: false,
                        willClose: false,
                        diff: null,
                        newBranch: boolean('branchChanged', false) ? 'the-new-branch-name' : 'the-old-branch-name',
                        oldBranch: 'the-old-branch-name',
                        branchChanged: boolean('branchChanged', false),
                    },
                ],
            },
        }}
    />
))

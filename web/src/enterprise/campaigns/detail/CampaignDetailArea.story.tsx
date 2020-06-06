import { storiesOf } from '@storybook/react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import React from 'react'
import { CampaignDetailArea } from './CampaignDetailArea'
import webStyles from '../../../SourcegraphWebApp.scss'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { of } from 'rxjs'
import * as H from 'history'
import { MemoryRouter } from 'react-router'

const history = H.createMemoryHistory()

const COMMON_PROPS: Pick<
    React.ComponentProps<typeof CampaignDetailArea>,
    'isLightTheme' | 'history' | 'location' | 'extensionsController' | 'platformContext' | 'telemetryService'
> = {
    isLightTheme: true,
    history,
    location: history.location,
    extensionsController: {} as any,
    platformContext: {} as any,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

const COMMON_CAMPAIGN_FIELDS: Pick<
    GQL.ICampaign,
    '__typename' | 'id' | 'name' | 'description' | 'url' | 'author' | 'branch' | 'createdAt' | 'updatedAt' | 'diffStat'
> = {
    __typename: 'Campaign' as const,
    id: 'c',
    name: 'My campaign',
    description: 'My description',
    url: 'https://example.com',
    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
    author: { username: 'alice' } as GQL.IUser,
    branch: 'b',
    createdAt: '2020-01-01',
    updatedAt: '2020-01-01',
    diffStat: {
        __typename: 'DiffStat' as const,
        added: 5,
        changed: 3,
        deleted: 2,
    },
}

export const SAMPLE_PatchConnection: GQL.IPatchConnection = {
    __typename: 'PatchConnection' as const,
    nodes: [
        {
            __typename: 'Patch' as const,
            id: 'p1',
            publicationEnqueued: false,
            diff: {
                fileDiffs: {
                    diffStat: { __typename: 'DiffStat' as const, added: 10, changed: 10, deleted: 10 },
                },
            },
            repository: { id: 'r', name: 'my/repo1', url: 'https://example.com' },
        },
    ] as GQL.IPatch[],
    pageInfo: { __typename: 'PageInfo' as const, hasNextPage: false, endCursor: null },
    totalCount: 1,
}

export const SAMPLE_FileDiffConnection: GQL.IFileDiffConnection = {
    __typename: 'FileDiffConnection' as const,
    nodes: [
        {
            __typename: 'FileDiff',
            hunks: [
                {
                    body: 'abc',
                    oldRange: { startLine: 1, lines: 2 },
                    newRange: { startLine: 1, lines: 2 },
                    highlight: {
                        lines: [
                            { html: 'diff line added', kind: GQL.DiffHunkLineType.ADDED },
                            { html: 'diff line deleted', kind: GQL.DiffHunkLineType.DELETED },
                        ] as GQL.IHighlightedDiffHunkLine[],
                    },
                },
            ],
            oldPath: 'foo.txt',
            oldFile: { __typename: 'VirtualFile', path: 'foo.txt' },
            newPath: 'foo.txt',
            newFile: { __typename: 'VirtualFile', path: 'foo.txt' },
            mostRelevantFile: { __typename: 'VirtualFile' },
            stat: { added: 5, changed: 5, deleted: 5 },
        },
    ] as GQL.IFileDiff[],
    pageInfo: { __typename: 'PageInfo' as const, hasNextPage: false, endCursor: null },
    totalCount: 1,
    diffStat: { __typename: 'DiffStat' as const, added: 10, changed: 10, deleted: 10 },
    rawDiff: 'abc',
}

export const SAMPLE_ExternalChangesetConnection: GQL.IExternalChangesetConnection = {
    __typename: 'ExternalChangesetConnection' as const,
    nodes: [
        {
            __typename: 'ExternalChangeset' as const,
            id: 'c1',
            title: 'My changeset 1',
            externalID: '1',
            labels: [
                { __typename: 'ChangesetLabel' as const, text: 'foo', color: 'blue', description: null },
            ] as GQL.IChangesetLabel[],
            repository: { name: 'my/repo1', url: 'https://example.com' },
            state: GQL.ChangesetState.OPEN,
            checkState: GQL.ChangesetCheckState.PASSED,
            reviewState: GQL.ChangesetReviewState.APPROVED,
            externalURL: {
                __typename: 'ExternalLink',
                url: 'https://example.com',
                serviceType: 'github',
            },
            createdAt: '2020-01-01',
            updatedAt: '2020-01-02',
        },
        {
            __typename: 'ExternalChangeset' as const,
            id: 'c2',
            title: 'My changeset 2',
            externalID: '2',
            labels: [
                { __typename: 'ChangesetLabel' as const, text: 'bar', color: 'blue', description: null },
            ] as GQL.IChangesetLabel[],
            repository: { name: 'my/repo2', url: 'https://example.com' },
            state: GQL.ChangesetState.OPEN,
            checkState: GQL.ChangesetCheckState.PENDING,
            reviewState: GQL.ChangesetReviewState.CHANGES_REQUESTED,
            externalURL: {
                __typename: 'ExternalLink',
                url: 'https://example.com',
                serviceType: 'github',
            },
            createdAt: '2020-01-02',
            updatedAt: '2020-01-03',
        },
    ] as GQL.IExternalChangeset[],
    pageInfo: { __typename: 'PageInfo' as const, hasNextPage: false, endCursor: null },
    totalCount: 2,
}

export const SAMPLE_PatchSet: GQL.IPatchSet = {
    __typename: 'PatchSet' as const,
    id: 'c',
    previewURL: 'https://example.com',
    diffStat: {
        __typename: 'DiffStat',
        added: 0,
        changed: 18,
        deleted: 999,
    },
    patches: {
        __typename: 'PatchConnection',
        nodes: [] as GQL.IPatch[],
        pageInfo: { __typename: 'PageInfo' as const, hasNextPage: false, endCursor: null },
        totalCount: 2,
    },
}

const SAMPLE_ChangesetCounts: GQL.IChangesetCounts[] = [
    { date: '2020-01-01', open: 3, merged: 0, openApproved: 0, total: 3 },
    { date: '2020-01-02', open: 2, merged: 1, openApproved: 1, openChangesRequested: 1, total: 3 },
    { date: '2020-01-03', open: 1, merged: 2, openApproved: 1, total: 3 },
    { date: '2020-01-04', open: 1, merged: 2, openApproved: 1, total: 4 },
    { date: '2020-01-05', open: 1, merged: 3, openApproved: 1, total: 4 },
    { date: '2020-01-06', open: 0, merged: 4, openApproved: 0, total: 4 },
] as GQL.IChangesetCounts[]

const { add } = storiesOf('web/campaigns/CampaignDetailArea', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light mt-3">{story()}</div>
    </>
))

add('With patches', () => (
    <MemoryRouter>
        <CampaignDetailArea
            {...COMMON_PROPS}
            authenticatedUser={{ id: 'u', username: 'alice', avatarURL: null }}
            campaign={{
                ...COMMON_CAMPAIGN_FIELDS,
                patchSet: { id: 'p' },
                changesets: { totalCount: 2 },
                openChangesets: { totalCount: 2 },
                patches: { totalCount: 1 },
                changesetCountsOverTime: SAMPLE_ChangesetCounts,
                hasUnpublishedPatches: true,
                viewerCanAdminister: true as boolean,
                status: {
                    completedCount: 0,
                    pendingCount: 0,
                    errors: [],
                    state: GQL.BackgroundProcessState.COMPLETED,
                },
                closedAt: null,
            }}
            fetchPatchSetById={() => of(SAMPLE_PatchSet)}
            queryPatchesFromCampaign={() => of(SAMPLE_PatchConnection)}
            queryPatchesFromPatchSet={() => of(SAMPLE_PatchConnection)}
            queryPatchFileDiffs={() => of(SAMPLE_FileDiffConnection)}
            queryChangesets={() => of(SAMPLE_ExternalChangesetConnection)}
        />
    </MemoryRouter>
))

add('Publishing', () => (
    <CampaignDetailArea
        {...COMMON_PROPS}
        authenticatedUser={{ id: 'u', username: 'alice', avatarURL: null }}
        campaign={{
            ...COMMON_CAMPAIGN_FIELDS,
            patchSet: { id: 'p' },
            changesets: { totalCount: 1 },
            openChangesets: { totalCount: 1 },
            patches: { totalCount: 1 },
            changesetCountsOverTime: SAMPLE_ChangesetCounts,
            hasUnpublishedPatches: true,
            viewerCanAdminister: true as boolean,
            status: {
                completedCount: 1,
                pendingCount: 1,
                errors: ['a'],
                state: GQL.BackgroundProcessState.PROCESSING,
            },
            closedAt: null,
        }}
        fetchPatchSetById={() => of(SAMPLE_PatchSet)}
        queryPatchesFromCampaign={() => of(SAMPLE_PatchConnection)}
        queryPatchesFromPatchSet={() => of(SAMPLE_PatchConnection)}
        queryPatchFileDiffs={() => of(SAMPLE_FileDiffConnection)}
        queryChangesets={() => of(SAMPLE_ExternalChangesetConnection)}
    />
))

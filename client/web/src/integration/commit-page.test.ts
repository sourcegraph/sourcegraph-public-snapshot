import { subDays } from 'date-fns'
import { afterEach, beforeEach, describe, it } from 'mocha'

import {
    DiffHunkLineType,
    ExternalServiceKind,
    RepositoryType,
    type SharedGraphQlOperations,
} from '@sourcegraph/shared/src/graphql-operations'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { createDriverForTest, type Driver } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { WebGraphQlOperations } from '../graphql-operations'

import { createWebIntegrationTestContext, type WebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

describe('RepositoryCommitPage', () => {
    const repositoryName = 'github.com/sourcegraph/sourcegraph'
    const commitID = '1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416'
    const commitDate = subDays(new Date(), 7).toISOString()
    const commonBlobGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
        ...commonWebGraphQlResults,
        RepositoryCommit: () => ({
            node: {
                __typename: 'Repository',
                sourceType: RepositoryType.GIT_REPOSITORY,
                commit: {
                    __typename: 'GitCommit',
                    id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3pOamd3T1RJMU1BPT0iLCJjIjoiMWU3YmQwMDBlNzhjZjM1YzZlMWJlMWI5ZjE1MTBiNGFhZGZhYTQxNiJ9',
                    oid: '1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416',
                    abbreviatedOID: '1e7bd00',
                    perforceChangelist: null,
                    message: 'Signup copy adjustment (#43435)\n\nCopy adjustment',
                    subject: 'Signup copy adjustment (#43435)',
                    body: 'Copy adjustment',
                    author: {
                        __typename: 'Signature',
                        person: {
                            __typename: 'Person',
                            avatarURL: null,
                            name: 'st0nebraker',
                            email: 'beccasteinbrecher@gmail.com',
                            displayName: 'Becca Steinbrecher',
                            user: {
                                __typename: 'User',
                                id: 'VXNlcjo1MTA5OQ==',
                                username: 'st0nebraker',
                                url: '/users/st0nebraker',
                                displayName: 'Becca Steinbrecher',
                            },
                        },
                        date: commitDate,
                    },
                    committer: {
                        __typename: 'Signature',
                        person: {
                            __typename: 'Person',
                            avatarURL: null,
                            name: 'GitHub',
                            email: 'noreply@github.com',
                            displayName: 'GitHub',
                            user: null,
                        },
                        date: commitDate,
                    },
                    parents: [
                        {
                            __typename: 'GitCommit',
                            oid: '56ab377d94fe96c87bc8c5e26675c585f9312e64',
                            abbreviatedOID: '56ab377',
                            perforceChangelist: null,
                            url: '/github.com/sourcegraph/sourcegraph/-/commit/56ab377d94fe96c87bc8c5e26675c585f9312e64',
                        },
                    ],
                    url: '/github.com/sourcegraph/sourcegraph/-/commit/1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416',
                    canonicalURL:
                        '/github.com/sourcegraph/sourcegraph/-/commit/1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416',
                    externalURLs: [
                        {
                            __typename: 'ExternalLink',
                            url: 'https://github.com/sourcegraph/sourcegraph/commit/1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416',
                            serviceKind: ExternalServiceKind.GITHUB,
                        },
                    ],
                    tree: {
                        __typename: 'GitTree',
                        canonicalURL: '/github.com/sourcegraph/sourcegraph@1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416',
                    },
                },
            },
        }),
        RepositoryComparisonDiff: () => ({
            node: {
                __typename: 'Repository',
                comparison: {
                    fileDiffs: {
                        nodes: [
                            {
                                oldPath: 'client/web/src/auth/CloudSignUpPage.tsx',
                                oldFile: {
                                    __typename: 'GitBlob',
                                    binary: false,
                                    byteSize: 8336,
                                },
                                newFile: {
                                    __typename: 'GitBlob',
                                    binary: false,
                                    byteSize: 8336,
                                },
                                newPath: 'client/web/src/auth/CloudSignUpPage.tsx',
                                mostRelevantFile: {
                                    __typename: 'GitBlob',
                                    url: '/github.com/sourcegraph/sourcegraph@1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416/-/blob/client/web/src/auth/CloudSignUpPage.tsx',
                                    changelistURL: '',
                                },
                                hunks: [
                                    {
                                        oldRange: {
                                            startLine: 167,
                                            lines: 7,
                                        },
                                        oldNoNewlineAt: false,
                                        newRange: {
                                            startLine: 167,
                                            lines: 7,
                                        },
                                        section:
                                            'export const CloudSignUpPage: React.FunctionComponent\u003CReact.PropsWithChildren\u003CPr',
                                        highlight: {
                                            aborted: false,
                                            lines: [
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                    \u003Cspan class="hl-punctuation hl-section hl-embedded hl-begin hl-tsx"\u003E{\u003C/span\u003E\u003Cspan class="hl-meta hl-embedded hl-expression hl-tsx"\u003E\u003Cspan class="hl-variable hl-other hl-readwrite hl-tsx"\u003EinvitedBy\u003C/span\u003E \u003Cspan class="hl-keyword hl-operator hl-ternary hl-tsx"\u003E?\u003C/span\u003E \u003Cspan class="hl-string hl-quoted hl-single hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-string hl-begin hl-tsx"\u003E\u0026#39;\u003C/span\u003EWith a Sourcegraph account, you can:\u003Cspan class="hl-punctuation hl-definition hl-string hl-end hl-tsx"\u003E\u0026#39;\u003C/span\u003E\u003C/span\u003E \u003Cspan class="hl-keyword hl-operator hl-ternary hl-tsx"\u003E:\u003C/span\u003E \u003Cspan class="hl-string hl-quoted hl-single hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-string hl-begin hl-tsx"\u003E\u0026#39;\u003C/span\u003EWith a Sourcegraph account, you can also:\u003Cspan class="hl-punctuation hl-definition hl-string hl-end hl-tsx"\u003E\u0026#39;\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-punctuation hl-section hl-embedded hl-end hl-tsx"\u003E}\u003C/span\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                    \u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eul\u003C/span\u003E\u003Cspan class="hl-meta hl-tag hl-attributes hl-tsx"\u003E \u003Cspan class="hl-entity hl-other hl-attribute-name hl-tsx"\u003EclassName\u003C/span\u003E\u003Cspan class="hl-keyword hl-operator hl-assignment hl-tsx"\u003E=\u003C/span\u003E\u003Cspan class="hl-punctuation hl-section hl-embedded hl-begin hl-tsx"\u003E{\u003C/span\u003E\u003Cspan class="hl-meta hl-embedded hl-expression hl-tsx"\u003E\u003Cspan class="hl-variable hl-other hl-object hl-tsx"\u003Estyles\u003C/span\u003E\u003Cspan class="hl-punctuation hl-accessor hl-tsx"\u003E.\u003C/span\u003E\u003Cspan class="hl-variable hl-other hl-property hl-tsx"\u003EfeatureList\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-punctuation hl-section hl-embedded hl-end hl-tsx"\u003E}\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.DELETED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                        \u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003ESearch across all your public repositories\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;/\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.ADDED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                        \u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003ESearch across 2M+ open source repositories\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;/\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                        \u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003EMonitor code for changes\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;/\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                        \u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003ENavigate through code with IDE like go to references and definition hovers\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;/\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-source hl-tsx"\u003E\u003Cspan class="hl-meta hl-var hl-expr hl-tsx"\u003E\u003Cspan class="hl-meta hl-arrow hl-tsx"\u003E\u003Cspan class="hl-meta hl-block hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E\u003Cspan class="hl-meta hl-tag hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003E                        \u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\u003Cspan class="hl-meta hl-tag hl-without-attributes hl-tsx"\u003E\u003Cspan class="hl-meta hl-jsx hl-children hl-tsx"\u003EIntegrate data, tooling, and code in a single location \u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-begin hl-tsx"\u003E\u0026lt;/\u003C/span\u003E\u003Cspan class="hl-entity hl-name hl-tag hl-tsx"\u003Eli\u003C/span\u003E\u003Cspan class="hl-punctuation hl-definition hl-tag hl-end hl-tsx"\u003E\u0026gt;\u003C/span\u003E\u003C/span\u003E\n\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/span\u003E\u003C/div\u003E',
                                                },
                                            ],
                                        },
                                    },
                                ],
                                stat: {
                                    added: 1,
                                    deleted: 1,
                                },
                                internalID: 'f931c927713915bc931dfd220bb90209',
                            },
                            {
                                oldPath: 'client/web/src/auth/__snapshots__/SignUpPage.test.tsx.snap',
                                oldFile: {
                                    __typename: 'GitBlob',
                                    binary: false,
                                    byteSize: 8630,
                                },
                                newFile: {
                                    __typename: 'GitBlob',
                                    binary: false,
                                    byteSize: 8630,
                                },
                                newPath: 'client/web/src/auth/__snapshots__/SignUpPage.test.tsx.snap',
                                mostRelevantFile: {
                                    __typename: 'GitBlob',
                                    url: '/github.com/sourcegraph/sourcegraph@1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416/-/blob/client/web/src/auth/__snapshots__/SignUpPage.test.tsx.snap',
                                    changelistURL: '',
                                },
                                hunks: [
                                    {
                                        oldRange: {
                                            startLine: 38,
                                            lines: 7,
                                        },
                                        oldNoNewlineAt: false,
                                        newRange: {
                                            startLine: 38,
                                            lines: 7,
                                        },
                                        section: 'exports[`SignUpPage renders sign up page (cloud) 1`] = `',
                                        highlight: {
                                            aborted: false,
                                            lines: [
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E          class=\u0026#34;featureList\u0026#34;\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E        \u0026gt;\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E          \u0026lt;li\u0026gt;\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.DELETED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E            Search across all your public repositories\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.ADDED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E            Search across 2M+ open source repositories\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E          \u0026lt;/li\u0026gt;\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E          \u0026lt;li\u0026gt;\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                                {
                                                    kind: DiffHunkLineType.UNCHANGED,
                                                    html: '\u003Cdiv\u003E\u003Cspan class="hl-text hl-plain"\u003E            Monitor code for changes\n\u003C/span\u003E\u003C/div\u003E',
                                                },
                                            ],
                                        },
                                    },
                                ],
                                stat: {
                                    added: 1,
                                    deleted: 1,
                                },
                                internalID: '150009a294c6bae96880a622085e5f07',
                            },
                        ],
                        totalCount: 2,
                        pageInfo: {
                            endCursor: null,
                            hasNextPage: false,
                        },
                        diffStat: {
                            __typename: 'DiffStat',
                            added: 2,
                            deleted: 2,
                        },
                    },
                },
            },
        }),
        ResolveRepoRev: () => ({
            repositoryRedirect: {
                __typename: 'Repository',
                id: 'UmVwb3NpdG9yeTozNjgwOTI1MA==',
                name: 'github.com/sourcegraph/sourcegraph',
                url: '/github.com/sourcegraph/sourcegraph',
                sourceType: RepositoryType.GIT_REPOSITORY,
                externalURLs: [
                    {
                        url: 'https://github.com/sourcegraph/sourcegraph',
                        serviceKind: ExternalServiceKind.GITHUB,
                    },
                ],
                externalRepository: {
                    serviceType: 'github',
                    serviceID: 'https://github.com/',
                },
                description: 'Universal code search (self-hosted)',
                viewerCanAdminister: false,
                defaultBranch: {
                    displayName: 'main',
                    abbrevName: 'main',
                },
                mirrorInfo: {
                    cloneInProgress: false,
                    cloneProgress: '',
                    cloned: true,
                },
                commit: {
                    __typename: 'GitCommit',
                    oid: '1e7bd000e78cf35c6e1be1b9f1510b4aadfaa416',
                    tree: {
                        url: '/github.com/sourcegraph/sourcegraph',
                    },
                },
                changelist: null,
                isFork: false,
                metadata: [],
            },
        }),
    }

    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('Display diff in unified mode', async () => {
        await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/commit/${commitID}`)
        await driver.page.waitForSelector('.test-file-diff-node', { visible: true })

        await percySnapshotWithVariants(driver.page, 'Commit page - Unified mode')
        await accessibilityAudit(driver.page)
    })

    it('Displays diff in split mode', async () => {
        await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/commit/${commitID}`)
        await driver.page.waitForSelector('.test-file-diff-node', { visible: true })

        const splitRadioButton = await driver.page.$('aria/Split[role="radio"]')
        await driver.page.evaluate(element => element.click(), splitRadioButton)

        await driver.page.waitForSelector('[data-split-mode="split"]', { visible: true })

        await percySnapshotWithVariants(driver.page, 'Commit page - Split mode')
        await accessibilityAudit(driver.page)
    })
})

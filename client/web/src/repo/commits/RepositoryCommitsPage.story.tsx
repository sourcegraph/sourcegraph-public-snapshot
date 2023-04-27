import { MockedResponse } from '@apollo/client/testing'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import { RepositoryFields, RepositoryGitCommitsResult, RepositoryGitCommitsVariables } from '../../graphql-operations'

import {
    REPOSITORY_GIT_COMMITS_QUERY,
    RepositoryCommitsPage,
    RepositoryCommitsPageProps,
} from './RepositoryCommitsPage'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/RepositoryCommitsPage',
    decorators: [decorator],
}

export default config

const mockRepositoryGitCommitsQuery: MockedResponse<RepositoryGitCommitsResult, RepositoryGitCommitsVariables> = {
    request: {
        query: getDocumentNode(REPOSITORY_GIT_COMMITS_QUERY),
        variables: {
            repo: 'UmVwb3NpdG9yeToyNjM4OQ==',
            revspec: '',
            filePath: '',
            first: 20,
            afterCursor: null,
        },
    },
    result: {
        data: {
            node: {
                isPerforceDepot: false,
                externalURLs: [
                    {
                        url: 'https://github.com/sourcegraph/sourcegraph',
                        serviceKind: 'GITHUB',
                        __typename: 'ExternalLink',
                    },
                ],
                commit: {
                    ancestors: {
                        nodes: [
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZmFiNmQ4MzQ0YTI4ZGVlNmFlMDlhOGNiMWFlNDk5NzNmMTE3MDU0YiJ9',
                                oid: 'fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                abbreviatedOID: 'fab6d83',
                                message:
                                    'batches: add changeset cell for displaying internal state (#50366)\n\nCo-authored-by: Kelli Rockwell \u003ckelli@sourcegraph.com\u003e\r',
                                subject: 'batches: add changeset cell for displaying internal state (#50366)',
                                body: 'Co-authored-by: Kelli Rockwell \u003ckelli@sourcegraph.com\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Bolaji Olajide',
                                        email: '25608335+BolajiOlajide@users.noreply.github.com',
                                        displayName: 'Bolaji Olajide',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:12:53Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:12:53Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'fb61a539c3a10075935ca78bf1334aa0260af040',
                                        abbreviatedOID: 'fb61a53',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@fab6d8344a28dee6ae09a8cb1ae49973f117054b',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZmI2MWE1MzljM2ExMDA3NTkzNWNhNzhiZjEzMzRhYTAyNjBhZjA0MCJ9',
                                oid: 'fb61a539c3a10075935ca78bf1334aa0260af040',
                                abbreviatedOID: 'fb61a53',
                                message:
                                    'Cody: ✨ Suggest follow-up topics (#51201)\n\nThis adds a new recipe that is used to suggest up to three follow-up\r\ntopics for Cody. The recipe is executed with every user chat message but\r\nwill not wait for the answer so that the experience is better (this is\r\nin line with comparable chat bots).\r\n\r\n## ToDo\r\n\r\n- [x] Add behind a feature flag for now \r\n\r\n## Test plan\r\n\r\n\r\n\r\nhttps://user-images.githubusercontent.com/458591/234884949-cd71893a-ee12-408f-8d7f-b6ca76497b66.mov\r\n\r\n\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e',
                                subject: 'Cody: ✨ Suggest follow-up topics (#51201)',
                                body: 'This adds a new recipe that is used to suggest up to three follow-up\r\ntopics for Cody. The recipe is executed with every user chat message but\r\nwill not wait for the answer so that the experience is better (this is\r\nin line with comparable chat bots).\r\n\r\n## ToDo\r\n\r\n- [x] Add behind a feature flag for now \r\n\r\n## Test plan\r\n\r\n\r\n\r\nhttps://user-images.githubusercontent.com/458591/234884949-cd71893a-ee12-408f-8d7f-b6ca76497b66.mov\r\n\r\n\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Philipp Spiess',
                                        email: 'hello@philippspiess.com',
                                        displayName: 'Philipp Spiess',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:02:51Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T16:02:51Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                        abbreviatedOID: '54605d8',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/fb61a539c3a10075935ca78bf1334aa0260af040',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@fb61a539c3a10075935ca78bf1334aa0260af040',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiNTQ2MDVkOGJkNDU2NjcwZTBiZDlkMmI0Mjk1ODUxYzAyNDBmY2Y5YyJ9',
                                oid: '54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                abbreviatedOID: '54605d8',
                                message:
                                    'Cody: Release 0.0.9 (#51206)\n\nPreparing for a new release by pushing the version number.\r\n\r\n## Test plan\r\n\r\nOnly a version number change.\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e',
                                subject: 'Cody: Release 0.0.9 (#51206)',
                                body: 'Preparing for a new release by pushing the version number.\r\n\r\n## Test plan\r\n\r\nOnly a version number change.\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Philipp Spiess',
                                        email: 'hello@philippspiess.com',
                                        displayName: 'Philipp Spiess',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T15:46:36Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T15:46:36Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '84dd723a59ee5e13b9bfed8357975357b36eb318',
                                        abbreviatedOID: '84dd723',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/84dd723a59ee5e13b9bfed8357975357b36eb318',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@54605d8bd456670e0bd9d2b4295851c0240fcf9c',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiODRkZDcyM2E1OWVlNWUxM2I5YmZlZDgzNTc5NzUzNTdiMzZlYjMxOCJ9',
                                oid: '84dd723a59ee5e13b9bfed8357975357b36eb318',
                                abbreviatedOID: '84dd723',
                                message:
                                    'app: setup exp: allow picking multiple folders (#50904)\n\nCloses #50833\r\n\r\nOn Mac, allows picking multiple folders with Cmd+click in the file\r\nbrowser in the setup wizard. This was done by modifying the AppleScript\r\nused to launch the file browsers; eventually, we probably want to make\r\nthis more system-agnostic by using the [Tauri file browse\r\ndialog](https://tauri.app/v1/api/js/dialog/) instead.\r\n\r\nThis also modifies the whole end-to-end to accept a list of paths,\r\nrather than a single string path, as the directory. Multiple code hosts\r\nwill be created, one for each path selected.\r\n\r\n\r\n## Test plan\r\n\r\nManually verify on the Tauri app:\r\n`sg start app` in one terminal, `pnpm tauri dev` in another.',
                                subject: 'app: setup exp: allow picking multiple folders (#50904)',
                                body: 'Closes #50833\r\n\r\nOn Mac, allows picking multiple folders with Cmd+click in the file\r\nbrowser in the setup wizard. This was done by modifying the AppleScript\r\nused to launch the file browsers; eventually, we probably want to make\r\nthis more system-agnostic by using the [Tauri file browse\r\ndialog](https://tauri.app/v1/api/js/dialog/) instead.\r\n\r\nThis also modifies the whole end-to-end to accept a list of paths,\r\nrather than a single string path, as the directory. Multiple code hosts\r\nwill be created, one for each path selected.\r\n\r\n\r\n## Test plan\r\n\r\nManually verify on the Tauri app:\r\n`sg start app` in one terminal, `pnpm tauri dev` in another.',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Juliana Peña',
                                        email: 'me@julip.co',
                                        displayName: 'Juliana Peña',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T15:25:06Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T15:25:06Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                        abbreviatedOID: '3e1679e',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/84dd723a59ee5e13b9bfed8357975357b36eb318',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/84dd723a59ee5e13b9bfed8357975357b36eb318',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/84dd723a59ee5e13b9bfed8357975357b36eb318',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@84dd723a59ee5e13b9bfed8357975357b36eb318',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiM2UxNjc5ZTBjYjA5ZmI1NjI5MWFjYmRkMTdmOGQ3ODAwMTIwZjBjMyJ9',
                                oid: '3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                abbreviatedOID: '3e1679e',
                                message:
                                    'grpc: buf: switch codeowners buf config to use non-deprecated plugin (#51057)\n',
                                subject:
                                    'grpc: buf: switch codeowners buf config to use non-deprecated plugin (#51057)',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Geoffrey Gilmore',
                                        email: 'geoffrey@sourcegraph.com',
                                        displayName: 'Geoffrey Gilmore',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T14:58:30Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T14:58:30Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '7933ebba5b1111a174e56100d81267aa8876a239',
                                        abbreviatedOID: '7933ebb',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/7933ebba5b1111a174e56100d81267aa8876a239',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@3e1679e0cb09fb56291acbdd17f8d7800120f0c3',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiNzkzM2ViYmE1YjExMTFhMTc0ZTU2MTAwZDgxMjY3YWE4ODc2YTIzOSJ9',
                                oid: '7933ebba5b1111a174e56100d81267aa8876a239',
                                abbreviatedOID: '7933ebb',
                                message:
                                    'completions: move things out of internal (#51189)\n\nCo-authored-by: davejrt \u003cdavetry@gmail.com\u003e',
                                subject: 'completions: move things out of internal (#51189)',
                                body: 'Co-authored-by: davejrt \u003cdavetry@gmail.com\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Joe Chen',
                                        email: 'joe@sourcegraph.com',
                                        displayName: 'Joe Chen',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T14:53:56Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T14:53:56Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                        abbreviatedOID: '006f187',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/7933ebba5b1111a174e56100d81267aa8876a239',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/7933ebba5b1111a174e56100d81267aa8876a239',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/7933ebba5b1111a174e56100d81267aa8876a239',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@7933ebba5b1111a174e56100d81267aa8876a239',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiMDA2ZjE4NzM1ZmQyY2NkMmNlMThlMWM5MDZiYWE2ZWQ2ODY2ZDIyYiJ9',
                                oid: '006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                abbreviatedOID: '006f187',
                                message:
                                    'Cody: add changelog (#51212)\n\nTrying to document all the changes we have made since last Friday for\r\nv0.0.9.\r\nFeel free to make changes directly to this PR if i missed anything!\r\n\r\n## Test plan\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e\r\n\r\ndocs change, test plan not required\r\n\r\n---------\r\n\r\nCo-authored-by: Erik Seliger \u003cerikseliger@me.com\u003e\r\nCo-authored-by: Philipp Spiess \u003chello@philippspiess.com\u003e',
                                subject: 'Cody: add changelog (#51212)',
                                body: 'Trying to document all the changes we have made since last Friday for\r\nv0.0.9.\r\nFeel free to make changes directly to this PR if i missed anything!\r\n\r\n## Test plan\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e\r\n\r\ndocs change, test plan not required\r\n\r\n---------\r\n\r\nCo-authored-by: Erik Seliger \u003cerikseliger@me.com\u003e\r\nCo-authored-by: Philipp Spiess \u003chello@philippspiess.com\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Beatrix',
                                        email: '68532117+abeatrix@users.noreply.github.com',
                                        displayName: 'Beatrix',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T14:03:26Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T14:03:26Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                        abbreviatedOID: '8c7bde3',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@006f18735fd2ccd2ce18e1c906baa6ed6866d22b',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiOGM3YmRlMzdlZTZkNzYzZDRkODk0N2FhOTAyY2FhNTYwYTA3MDFjMSJ9',
                                oid: '8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                abbreviatedOID: '8c7bde3',
                                message:
                                    'autoindexing: move to scip-go from lsif-go (#50407)\n\n## Test plan\r\n\r\n- [ ] Test out that we still get precise code intel on several different\r\nrepos\r\n- [ ] Test out that this properly uploads to our different instances\r\n\r\n---------\r\n\r\nCo-authored-by: Jean-Hadrien Chabran \u003cjh@chabran.fr\u003e',
                                subject: 'autoindexing: move to scip-go from lsif-go (#50407)',
                                body: '## Test plan\r\n\r\n- [ ] Test out that we still get precise code intel on several different\r\nrepos\r\n- [ ] Test out that this properly uploads to our different instances\r\n\r\n---------\r\n\r\nCo-authored-by: Jean-Hadrien Chabran \u003cjh@chabran.fr\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'TJ DeVries',
                                        email: 'devries.timothyj@gmail.com',
                                        displayName: 'TJ DeVries',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T13:09:31Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T13:09:31Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                        abbreviatedOID: '4ef01c4',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@8c7bde37ee6d763d4d8947aa902caa560a0701c1',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiNGVmMDFjNDczMDA0M2E4NzY0YWVlMDM4ZGUyZmRjZDBhNTJhNmZhYyJ9',
                                oid: '4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                abbreviatedOID: '4ef01c4',
                                message:
                                    '(feat) add repo metadata editing UI (#50821)\n\n* update repo metadata display UI on the search results and repo root pages according to new designs\r\n\r\nCo-authored-by: Tim Lucas \u003ct@toolmantim.com\u003e',
                                subject: '(feat) add repo metadata editing UI (#50821)',
                                body: '* update repo metadata display UI on the search results and repo root pages according to new designs\r\n\r\nCo-authored-by: Tim Lucas \u003ct@toolmantim.com\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Erzhan Torokulov',
                                        email: 'erzhan.torokulov@gmail.com',
                                        displayName: 'Erzhan Torokulov',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T12:48:06Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T12:48:06Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                        abbreviatedOID: '08c0bf1',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@4ef01c4730043a8764aee038de2fdcd0a52a6fac',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiMDhjMGJmMWUzM2NlMGIyZGUwZmNiNmJiOWRiMjYzZTUyZDE3MWI3ZCJ9',
                                oid: '08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                abbreviatedOID: '08c0bf1',
                                message:
                                    'Cody: Recent changes recipes message (#51098)\n\nThis PR is for issue #51084. There was a console message when no git log\r\nwas present for the selected dropdown. This PR implemented the\r\nfunctionality to return a response as "No recent changes found".\r\n\r\n## Test plan\r\n\r\nAll the manual and integration tests have passed.\r\nManual testing has been done to check the expected behaviour.',
                                subject: 'Cody: Recent changes recipes message (#51098)',
                                body: 'This PR is for issue #51084. There was a console message when no git log\r\nwas present for the selected dropdown. This PR implemented the\r\nfunctionality to return a response as "No recent changes found".\r\n\r\n## Test plan\r\n\r\nAll the manual and integration tests have passed.\r\nManual testing has been done to check the expected behaviour.',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Deepak Kumar',
                                        email: 'deepakdk2431@gmail.com',
                                        displayName: 'Deepak Kumar',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T12:12:44Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T12:12:44Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '3f068773e79e106de8cad881d86c732090f611bb',
                                        abbreviatedOID: '3f06877',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/3f068773e79e106de8cad881d86c732090f611bb',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@08c0bf1e33ce0b2de0fcb6bb9db263e52d171b7d',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiM2YwNjg3NzNlNzllMTA2ZGU4Y2FkODgxZDg2YzczMjA5MGY2MTFiYiJ9',
                                oid: '3f068773e79e106de8cad881d86c732090f611bb',
                                abbreviatedOID: '3f06877',
                                message:
                                    'Cody Web: Error handling, especially for rate limits (#51204)\n\nCloses #51090\r\n\r\n## Test plan\r\n\r\n\u003cimg width="586" alt="Screenshot 2023-04-27 at 13 16 20"\r\nsrc="https://user-images.githubusercontent.com/458591/234846786-330bfd10-1f53-435a-a97a-36327d7b78dd.png"\u003e\r\n\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e',
                                subject: 'Cody Web: Error handling, especially for rate limits (#51204)',
                                body: 'Closes #51090\r\n\r\n## Test plan\r\n\r\n\u003cimg width="586" alt="Screenshot 2023-04-27 at 13 16 20"\r\nsrc="https://user-images.githubusercontent.com/458591/234846786-330bfd10-1f53-435a-a97a-36327d7b78dd.png"\u003e\r\n\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Philipp Spiess',
                                        email: 'hello@philippspiess.com',
                                        displayName: 'Philipp Spiess',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T12:06:20Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T12:06:20Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'd304f6173c0ece93227baba107d3247f942582d7',
                                        abbreviatedOID: 'd304f61',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/d304f6173c0ece93227baba107d3247f942582d7',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/3f068773e79e106de8cad881d86c732090f611bb',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/3f068773e79e106de8cad881d86c732090f611bb',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/3f068773e79e106de8cad881d86c732090f611bb',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@3f068773e79e106de8cad881d86c732090f611bb',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZDMwNGY2MTczYzBlY2U5MzIyN2JhYmExMDdkMzI0N2Y5NDI1ODJkNyJ9',
                                oid: 'd304f6173c0ece93227baba107d3247f942582d7',
                                abbreviatedOID: 'd304f61',
                                message:
                                    'Cody settings trailing slash fixed (#51163)\n\nThis PR is for issue #51155. While adding the codebase name with a\r\ntrailing "/" it was showing as an error. This PR fixed this bug, now it\r\naccepts codebase name either with "/" or not with it.\r\n\r\n\r\n## Test plan\r\nAll the unit and integration tests have been passed.\r\nTested the behaviour manually it works fine.',
                                subject: 'Cody settings trailing slash fixed (#51163)',
                                body: 'This PR is for issue #51155. While adding the codebase name with a\r\ntrailing "/" it was showing as an error. This PR fixed this bug, now it\r\naccepts codebase name either with "/" or not with it.\r\n\r\n\r\n## Test plan\r\nAll the unit and integration tests have been passed.\r\nTested the behaviour manually it works fine.',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Deepak Kumar',
                                        email: 'deepakdk2431@gmail.com',
                                        displayName: 'Deepak Kumar',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T11:39:14Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T11:39:14Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                        abbreviatedOID: '8a3bc68',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/d304f6173c0ece93227baba107d3247f942582d7',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/d304f6173c0ece93227baba107d3247f942582d7',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/d304f6173c0ece93227baba107d3247f942582d7',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@d304f6173c0ece93227baba107d3247f942582d7',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiOGEzYmM2ODA4YjgxM2ZjNTliMmY4MjljNmE4YmQ5ZjlkZjFiZWQxMCJ9',
                                oid: '8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                abbreviatedOID: '8a3bc68',
                                message:
                                    "completions: Switch to daily window for rate limiting (#51200)\n\nAfter discussing the limits yesterday, we've decided to go with a daily\r\ncap vs a lower hourly cap to help most users have a productive,\r\nuninterrupted experience with cody.",
                                subject: 'completions: Switch to daily window for rate limiting (#51200)',
                                body: "After discussing the limits yesterday, we've decided to go with a daily\r\ncap vs a lower hourly cap to help most users have a productive,\r\nuninterrupted experience with cody.",
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Erik Seliger',
                                        email: 'erikseliger@me.com',
                                        displayName: 'Erik Seliger',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T11:13:02Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T11:13:02Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                        abbreviatedOID: '4cc162d',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@8a3bc6808b813fc59b2f829c6a8bd9f9df1bed10',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiNGNjMTYyZDI2NTQyMzhhNTJjZTBjNDUyNTcwMjhjYmM4NGE4Yjk1ZCJ9',
                                oid: '4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                abbreviatedOID: '4cc162d',
                                message:
                                    'chore: update changelog with fix for search contexts (#51203)\n\n## Test plan\r\n\r\nNo testing, just changelog update 🤷',
                                subject: 'chore: update changelog with fix for search contexts (#51203)',
                                body: '## Test plan\r\n\r\nNo testing, just changelog update 🤷',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Milan Freml',
                                        email: 'kopancek@users.noreply.github.com',
                                        displayName: 'Milan Freml',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T10:16:01Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T10:16:01Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '31c9497b78a72552a0cf4633b22873f70a823b15',
                                        abbreviatedOID: '31c9497',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/31c9497b78a72552a0cf4633b22873f70a823b15',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@4cc162d2654238a52ce0c45257028cbc84a8b95d',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiMzFjOTQ5N2I3OGE3MjU1MmEwY2Y0NjMzYjIyODczZjcwYTgyM2IxNSJ9',
                                oid: '31c9497b78a72552a0cf4633b22873f70a823b15',
                                abbreviatedOID: '31c9497',
                                message:
                                    "Cody: Add jaccard-based context to inline completions (#51161)\n\nThis brings the inline completion suggestions quality up to parity with\r\nmultiline suggestions by adding the same jaccard-based context based on\r\nopen editor tabs and the editor history into the prompt.\r\n\r\nTo avoid the two implementation diverging again in the future, I've also\r\nput some of the logic into an `abstract class`. Unfortunately the diff\r\nis hard to read because I also moved out the completion providers into\r\ntheir own file now.\r\n\r\nBehavior for multiline completions should be unchanged.\r\n\r\nWhile working on this, I had a few really good inline completions from\r\nCody already 🎉 This is starting to get somewhere!\r\n\r\n## Test plan\r\n\r\nI added a bunch of `console.log`s to the prompt creation code to make\r\nsure the snippets that are added to the knowledge library are\r\nreasonable. This is getting really annoying to debug, so a better way to\r\nlog which prompts are created is starting to get more interesting IMO.\r\n\r\n\r\n\r\nhttps://user-images.githubusercontent.com/458591/234633770-7a703720-84a9-43f2-bea3-07c2715c07c0.mov\r\n\r\n\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e",
                                subject: 'Cody: Add jaccard-based context to inline completions (#51161)',
                                body: "This brings the inline completion suggestions quality up to parity with\r\nmultiline suggestions by adding the same jaccard-based context based on\r\nopen editor tabs and the editor history into the prompt.\r\n\r\nTo avoid the two implementation diverging again in the future, I've also\r\nput some of the logic into an `abstract class`. Unfortunately the diff\r\nis hard to read because I also moved out the completion providers into\r\ntheir own file now.\r\n\r\nBehavior for multiline completions should be unchanged.\r\n\r\nWhile working on this, I had a few really good inline completions from\r\nCody already 🎉 This is starting to get somewhere!\r\n\r\n## Test plan\r\n\r\nI added a bunch of `console.log`s to the prompt creation code to make\r\nsure the snippets that are added to the knowledge library are\r\nreasonable. This is getting really annoying to debug, so a better way to\r\nlog which prompts are created is starting to get more interesting IMO.\r\n\r\n\r\n\r\nhttps://user-images.githubusercontent.com/458591/234633770-7a703720-84a9-43f2-bea3-07c2715c07c0.mov\r\n\r\n\r\n\r\n\u003c!-- All pull requests REQUIRE a test plan:\r\nhttps://docs.sourcegraph.com/dev/background-information/testing_principles\r\n--\u003e",
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Philipp Spiess',
                                        email: 'hello@philippspiess.com',
                                        displayName: 'Philipp Spiess',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T10:01:51Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T10:01:51Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '128a0b977541d5dd6d3e62d81013fee2b9075401',
                                        abbreviatedOID: '128a0b9',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/128a0b977541d5dd6d3e62d81013fee2b9075401',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/31c9497b78a72552a0cf4633b22873f70a823b15',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/31c9497b78a72552a0cf4633b22873f70a823b15',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/31c9497b78a72552a0cf4633b22873f70a823b15',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@31c9497b78a72552a0cf4633b22873f70a823b15',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiMTI4YTBiOTc3NTQxZDVkZDZkM2U2MmQ4MTAxM2ZlZTJiOTA3NTQwMSJ9',
                                oid: '128a0b977541d5dd6d3e62d81013fee2b9075401',
                                abbreviatedOID: '128a0b9',
                                message:
                                    'fix: edit search context with special characters results in 404 (#51196)\n\nEscaped the path parameter to stop this from happening.\r\n\r\nfixes #50992 \r\n- #50992\r\n\r\n## Test plan\r\n\r\nTested locally that it works as expected.\r\n\r\nNot sure if we want to extend the existing unit tests or work on a\r\ndifferent fix.',
                                subject: 'fix: edit search context with special characters results in 404 (#51196)',
                                body: 'Escaped the path parameter to stop this from happening.\r\n\r\nfixes #50992 \r\n- #50992\r\n\r\n## Test plan\r\n\r\nTested locally that it works as expected.\r\n\r\nNot sure if we want to extend the existing unit tests or work on a\r\ndifferent fix.',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Milan Freml',
                                        email: 'kopancek@users.noreply.github.com',
                                        displayName: 'Milan Freml',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T09:53:56Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T09:53:56Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'faf80a25683ad0db4c59caccbb68918602731852',
                                        abbreviatedOID: 'faf80a2',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/faf80a25683ad0db4c59caccbb68918602731852',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/128a0b977541d5dd6d3e62d81013fee2b9075401',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/128a0b977541d5dd6d3e62d81013fee2b9075401',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/128a0b977541d5dd6d3e62d81013fee2b9075401',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@128a0b977541d5dd6d3e62d81013fee2b9075401',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZmFmODBhMjU2ODNhZDBkYjRjNTljYWNjYmI2ODkxODYwMjczMTg1MiJ9',
                                oid: 'faf80a25683ad0db4c59caccbb68918602731852',
                                abbreviatedOID: 'faf80a2',
                                message:
                                    'dev: add basic embeddings QA script (#51089)\n\nThis adds the "sg embeddings-qa" command. The command calculates\r\n_recall_ for a curated test set (prompts -\u003e relevant context files). The\r\ngoal is to use recall to assess changes to the embeddings service. The\r\nlist of prompts will grow over time.\r\n\r\nCo-authored-by: Julie Tibshirani \u003cjulie.tibshirani@sourcegraph.com\u003e',
                                subject: 'dev: add basic embeddings QA script (#51089)',
                                body: 'This adds the "sg embeddings-qa" command. The command calculates\r\n_recall_ for a curated test set (prompts -\u003e relevant context files). The\r\ngoal is to use recall to assess changes to the embeddings service. The\r\nlist of prompts will grow over time.\r\n\r\nCo-authored-by: Julie Tibshirani \u003cjulie.tibshirani@sourcegraph.com\u003e',
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Stefan Hengl',
                                        email: 'stefan@sourcegraph.com',
                                        displayName: 'Stefan Hengl',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T08:09:27Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T08:09:27Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                        abbreviatedOID: 'aae1d66',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/faf80a25683ad0db4c59caccbb68918602731852',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/faf80a25683ad0db4c59caccbb68918602731852',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/faf80a25683ad0db4c59caccbb68918602731852',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@faf80a25683ad0db4c59caccbb68918602731852',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiYWFlMWQ2NjcyMDNmYmQzN2I0ZmIxZjEzZmM5YzdiMzY5YzkwMjQ3OCJ9',
                                oid: 'aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                abbreviatedOID: 'aae1d66',
                                message: '[github app] Add redis cache for installation access tokens (#51092)\n',
                                subject: '[github app] Add redis cache for installation access tokens (#51092)',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Petri-Johan Last',
                                        email: 'petri.last@sourcegraph.com',
                                        displayName: 'Petri-Johan Last',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T08:03:13Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-27T08:03:13Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'd8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                        abbreviatedOID: 'd8f53d2',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/d8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@aae1d667203fbd37b4fb1f13fc9c7b369c902478',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZDhmNTNkMmNiYTY3ODY0Y2ZkZThlNTZkODJjZDk2YzBiMzY5OTc5MSJ9',
                                oid: 'd8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                abbreviatedOID: 'd8f53d2',
                                message: 'ranking: Fix transaction safety when exporting uploads (#51168)\n',
                                subject: 'ranking: Fix transaction safety when exporting uploads (#51168)',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Eric Fritz',
                                        email: 'eric@sourcegraph.com',
                                        displayName: 'Eric Fritz',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-26T23:58:38Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-26T23:58:38Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                        abbreviatedOID: 'f06dd2b',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/d8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/d8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/d8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@d8f53d2cba67864cfde8e56d82cd96c0b3699791',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUbzMiLCJjIjoiZjA2ZGQyYjI1MjMyYmJiNTJmYjRiYTAxNWU4ZDE5Yzg2ZWVjYTFiYSJ9',
                                oid: 'f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                abbreviatedOID: 'f06dd2b',
                                message: 'ranking: Fix fk constraint error on insertion in mapper jobs (#51192)\n',
                                subject: 'ranking: Fix fk constraint error on insertion in mapper jobs (#51192)',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'Eric Fritz',
                                        email: 'eric@sourcegraph.com',
                                        displayName: 'Eric Fritz',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-26T23:53:58Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'GitHub',
                                        email: 'noreply@github.com',
                                        displayName: 'GitHub',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2023-04-26T23:53:58Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: '422f246832927af2ca100d2b85aed418e5fbdf9a',
                                        abbreviatedOID: '422f246',
                                        url: '/github.com/sourcegraph/sourcegraph/-/commit/422f246832927af2ca100d2b85aed418e5fbdf9a',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/github.com/sourcegraph/sourcegraph/-/commit/f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                canonicalURL:
                                    '/github.com/sourcegraph/sourcegraph/-/commit/f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                externalURLs: [
                                    {
                                        url: 'https://github.com/sourcegraph/sourcegraph/commit/f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                        serviceKind: 'GITHUB',
                                        __typename: 'ExternalLink',
                                    },
                                ],
                                tree: {
                                    canonicalURL:
                                        '/github.com/sourcegraph/sourcegraph@f06dd2b25232bbb52fb4ba015e8d19c86eeca1ba',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                        ],
                        pageInfo: { hasNextPage: true, endCursor: '20', __typename: 'PageInfo' },
                        __typename: 'GitCommitConnection',
                    },
                    __typename: 'GitCommit',
                },
                __typename: 'Repository',
            },
        },
    },
}

const mockRepositoryPerforceChangelistsQuery: MockedResponse<
    RepositoryGitCommitsResult,
    RepositoryGitCommitsVariables
> = {
    request: {
        query: getDocumentNode(REPOSITORY_GIT_COMMITS_QUERY),
        variables: {
            repo: 'UmVwb3NpdG9yeToyNjM4OQ==',
            revspec: '',
            filePath: '',
            first: 20,
            afterCursor: null,
        },
    },
    result: {
        data: {
            node: {
                isPerforceDepot: true,
                externalURLs: [],
                commit: {
                    ancestors: {
                        nodes: [
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJhNzdkMWIzMDliY2U1ZGJiNjM1ODFiY2E4YzJjYWFjMDEzZWM5Mzg3In0=',
                                oid: '48485',
                                abbreviatedOID: '48485',
                                message: '48485 - test-5386\n[p4-fusion: depot-paths = "//go/": change = 48485]',
                                subject: 'test-5386',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-13T19:24:59Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-13T19:24:59Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'd9994dc548fd79b473ce05198c88282890983fa9',
                                        abbreviatedOID: 'd9994dc',
                                        url: '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/perforce.sgdev.org/go/-/commit/a77d1b309bce5dbb63581bca8c2caac013ec9387',
                                canonicalURL:
                                    '/perforce.sgdev.org/go/-/commit/a77d1b309bce5dbb63581bca8c2caac013ec9387',
                                externalURLs: [],
                                tree: {
                                    canonicalURL: '/perforce.sgdev.org/go@a77d1b309bce5dbb63581bca8c2caac013ec9387',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJkOTk5NGRjNTQ4ZmQ3OWI0NzNjZTA1MTk4Yzg4MjgyODkwOTgzZmE5In0=',
                                oid: '1012',
                                abbreviatedOID: '1012',
                                message: '1012 - :boar:\n\n[p4-fusion: depot-paths = "//go/": change = 1012]',
                                subject: ':boar:',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T22:01:07Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T22:01:07Z',
                                    __typename: 'Signature',
                                },
                                parents: [
                                    {
                                        oid: 'd7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                        abbreviatedOID: 'd7205eb',
                                        url: '/perforce.sgdev.org/go/-/commit/d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                        __typename: 'GitCommit',
                                    },
                                ],
                                url: '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
                                canonicalURL:
                                    '/perforce.sgdev.org/go/-/commit/d9994dc548fd79b473ce05198c88282890983fa9',
                                externalURLs: [],
                                tree: {
                                    canonicalURL: '/perforce.sgdev.org/go@d9994dc548fd79b473ce05198c88282890983fa9',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                            {
                                id: 'R2l0Q29tbWl0OnsiciI6IlVtVndiM05wZEc5eWVUb3lOak00T1E9PSIsImMiOiJkNzIwNWViZjUxYzJjNzQ4NmVlNzRkZWQ5ZjAwMGRkZjFiMGNjYTI0In0=',
                                oid: '1011',
                                abbreviatedOID: '1011',
                                message:
                                    '1011 - Add Go source code\n\n[p4-fusion: depot-paths = "//go/": change = 1011]',
                                subject: 'Add Go source code',
                                body: null,
                                author: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T21:44:50Z',
                                    __typename: 'Signature',
                                },
                                committer: {
                                    person: {
                                        avatarURL: null,
                                        name: 'admin',
                                        email: 'admin@perforce-server-7df6ff678c-lkzfb',
                                        displayName: 'admin',
                                        user: null,
                                        __typename: 'Person',
                                    },
                                    date: '2021-08-09T21:44:50Z',
                                    __typename: 'Signature',
                                },
                                parents: [],
                                url: '/perforce.sgdev.org/go/-/commit/d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                canonicalURL:
                                    '/perforce.sgdev.org/go/-/commit/d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                externalURLs: [],
                                tree: {
                                    canonicalURL: '/perforce.sgdev.org/go@d7205ebf51c2c7486ee74ded9f000ddf1b0cca24',
                                    __typename: 'GitTree',
                                },
                                __typename: 'GitCommit',
                            },
                        ],
                        pageInfo: { hasNextPage: false, endCursor: null, __typename: 'PageInfo' },
                        __typename: 'GitCommitConnection',
                    },
                    __typename: 'GitCommit',
                },
                __typename: 'Repository',
            },
        },
    },
}

const repo: RepositoryFields = {
    id: 'repo-id',
    name: 'github.com/sourcegraph/sourcegraph',
    url: 'https://github.com/sourcegraph/sourcegraph/perforce',
    isPerforceDepot: false,
    description: '',
    viewerCanAdminister: false,
    isFork: false,
    externalURLs: [],
    externalRepository: {
        __typename: 'ExternalRepository',
        serviceType: '',
        serviceID: '',
    },
    defaultBranch: null,
    metadata: [],
}

export const GitCommitsStory: Story<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/github.com/sourcegraph/sourcegraph/-/commits']}
            mocks={[mockRepositoryGitCommitsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={repo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

GitCommitsStory.storyName = 'Git commits'

export const PerforceChangelistsStory: Story<RepositoryCommitsPageProps> = () => (
    <MockedTestProvider>
        <WebStory
            initialEntries={['/perforce.sgdev.org/go/-/commits']}
            mocks={[mockRepositoryPerforceChangelistsQuery]}
        >
            {props => <RepositoryCommitsPage revision="" repo={repo} {...props} />}
        </WebStory>
    </MockedTestProvider>
)

PerforceChangelistsStory.storyName = 'Perforce changelists'

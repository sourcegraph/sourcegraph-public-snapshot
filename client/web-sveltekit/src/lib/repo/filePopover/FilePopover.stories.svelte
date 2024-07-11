<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'

    import type { FilePopoverFragment, DirPopoverFragment, FilePopoverLastCommitFragment } from './FilePopover.gql'
    import FilePopover from './FilePopover.svelte'

    export const meta = {
        component: FilePopover,
    }

    const lastCommit: FilePopoverLastCommitFragment = {
        abbreviatedOID: '1234567',
        oid: '1234567890123456789012345678901234567890',
        subject: 'Test subject',
        canonicalURL: 'https://sourcegraph.com/about',
        author: {
            date: '2021-01-01T00:00:00Z',
            person: {
                __typename: 'Person',
                name: 'camdencheek',
                avatarURL: 'https://github.com/camdencheek.png',
                displayName: 'Camden Cheek',
            },
        },
    }

    const fileEntry: FilePopoverFragment = {
        __typename: 'GitBlob',
        name: 'results.go',
        path: 'internal/search/results/results.go',
        languages: ['Go'],
        byteSize: 325,
        totalLines: 12,
        history: {
            nodes: [
                {
                    commit: lastCommit,
                },
            ],
        },
    }

    const dirEntry: DirPopoverFragment = {
        __typename: 'GitTree',
        path: 'internal/search/results',
        files: [{ name: 'results.go' }, { name: 'results_test.go' }],
        directories: [{ name: 'testdata' }],
        history: {
            nodes: [{ commit: lastCommit }],
        },
    }
</script>

<Story name="Default">
    <FilePopover repoName={'github.com/sourcegraph/sourcegraph'} entry={fileEntry} />

    <FilePopover repoName={'github.com/sourcegraph/sourcegraph'} entry={dirEntry} />
</Story>

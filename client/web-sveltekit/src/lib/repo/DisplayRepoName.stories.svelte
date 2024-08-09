<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'
    import type { ComponentProps } from 'svelte'

    import { highlightRanges } from '$lib/dom'
    import { ExternalServiceKind } from '$lib/graphql-types'

    import DisplayRepoName from './DisplayRepoName.svelte'

    export const meta = {
        component: DisplayRepoName,
    }
</script>

<script lang="ts">
    const cases: ComponentProps<DisplayRepoName>[] = [
        { repoName: 'github.com/sourcegraph/sourcegraph', kind: undefined },
        { repoName: 'github.com/sourcegraph/sourcegraph', kind: ExternalServiceKind.GITHUB },
        { repoName: 'github.com/sourcegraph/sourcegraph', kind: ExternalServiceKind.GITLAB },
        { repoName: 'bitbucket.com/sourcegraph/sourcegraph', kind: ExternalServiceKind.BITBUCKETCLOUD },
        { repoName: 'bitbucket.com/sourcegraph/sourcegraph', kind: ExternalServiceKind.BITBUCKETSERVER },
        { repoName: 'bitbucket.com/sourcegraph/sourcegraph', kind: undefined },
        { repoName: 'gitlab.com/sourcegraph/sourcegraph', kind: undefined },
        { repoName: 'gitlab.com/sourcegraph/sourcegraph', kind: ExternalServiceKind.GITLAB },
        { repoName: 'mytestrepo', kind: undefined },
        { repoName: 'mytestrepo', kind: ExternalServiceKind.GITHUB },
        { repoName: 'ghe.sgdev.org/sourcegraph/sourcegraph', kind: undefined },

        // TODO(camdencheek): This is a scenario where, if we knew the external URL,
        // we could trim the hostname and put it in the tooltip. We can't safely do
        // this right now because an admin can name their repos whatever they want,
        // so we can't guarantee that the first element of a repo name is the host name
        // of the code host.
        { repoName: 'ghe.sgdev.org/sourcegraph/sourcegraph', kind: ExternalServiceKind.GITHUB },
    ]
</script>

<Story name="Default">
    <h2>DisplayRepoName props</h2>
    <table>
        {#each cases as testCase}
            <tr>
                <th>{testCase.repoName}</th>
                <th>{testCase.kind}</th>
                <td><div><DisplayRepoName {...testCase} /><div /></div></td>
            </tr>
        {/each}
    </table>
</Story>

<Story name="Highlighting works">
    <h2>"Graph" should be highlighted</h2>
    <span use:highlightRanges={{ ranges: [[17, 22]] }}
        ><DisplayRepoName repoName="github.com/sourcegraph/conc" kind={undefined} /></span
    >
</Story>

<style lang="scss">
    td,
    th {
        padding: 0.5rem;
    }

    td > div {
        width: min-content;
        border: 1px solid black;
    }
</style>

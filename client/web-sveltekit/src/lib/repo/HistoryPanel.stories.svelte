<script lang="ts" context="module">
    import { createHistoryResults } from '$testing/testdata'
    import { Story } from '@storybook/addon-svelte-csf'
    import HistoryPanel from './HistoryPanel.svelte'
    import { readable } from 'svelte/store'
    export const meta = {
        component: HistoryPanel,
        parameters: {
            sveltekit_experimental: {
                stores: {
                    page: {
                        url: new URL(window.location.href),
                    },
                },
            },
        },
    }
</script>

<script lang="ts">
    let commitCount = 5
    $: [initial] = createHistoryResults(1, commitCount)
    $: store = {
        ...readable({ data: initial.nodes, fetching: false }),
        fetchMore: () => {},
        fetchWhile: () => Promise.resolve(),
        capture: () => undefined,
        restore: () => Promise.resolve(),
    }
</script>

<Story name="Default">
    <p>Commits to show: <input type="number" bind:value={commitCount} min="1" max="100" /></p>
    <hr />
    {#key commitCount}
        <HistoryPanel history={store} enableInlineDiff={false} />
    {/key}
</Story>

<script lang="ts" context="module">
    import { createHistoryResults } from '$testing/testdata'
    import { Story } from '@storybook/addon-svelte-csf'
    import HistoryPanel from './HistoryPanel.svelte'
    export const meta = {
        component: HistoryPanel,
        parameters: {
            sveltekit_experimental: {
                stores: {
                    page: {},
                },
            },
        },
    }
</script>

<script lang="ts">
    let commitCount = 5
    $: [initial] = createHistoryResults(1, commitCount)
</script>

<Story name="Default">
    <p>Commits to show: <input type="number" bind:value={commitCount} min="1" max="100" /></p>
    <hr />
    {#key commitCount}
        <HistoryPanel history={initial} enableInlineDiffs={false} fetchMore={() => {}} />
    {/key}
</Story>

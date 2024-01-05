<script lang="ts" context="module">
    import { createHistoryResults } from '$testdata'
    import { Story } from '@storybook/addon-svelte-csf'
    import HistoryPanel from './HistoryPanel.svelte'
    export const meta = {
        component: HistoryPanel,
    }
</script>

<script lang="ts">
    let commitCount = 5
    $: [initial, next] = createHistoryResults(2, commitCount)
</script>

<Story name="Default">
    <p>Commits to show: <input type="number" bind:value={commitCount} min="1" max="100" /></p>
    <hr />
    {#key commitCount}
        <HistoryPanel history={Promise.resolve(initial)} fetchMoreHandler={async () => next} />
    {/key}
</Story>

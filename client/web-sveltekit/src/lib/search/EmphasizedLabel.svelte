<script lang="ts">
    import { getSpans } from '$lib/branded'

    export let label: string
    export let matches: Set<number> | null = null
    export let offset: number = 0

    $: spans = matches ? getSpans(matches, label.length, offset) : null
</script>

{#if spans}
    {#each spans as [start, end, match]}
        {#if match}
            <span class:match>{label.slice(start, end + 1)}</span>
        {:else}
            {label.slice(start, end + 1)}
        {/if}
    {/each}
{:else}
    {label}
{/if}

<style lang="scss">
    .match {
        font-weight: bold;
    }
</style>

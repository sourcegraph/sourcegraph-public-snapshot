<script lang="ts">
    import { decorateQuery } from '$lib/branded'

    export let query: string

    $: decorations = decorateQuery(query)
</script>

<span class="text-monospace search-query-link">
    {#if decorations}
        {#each decorations as { key, className, value } (key)}
            <span class={className}>{value}</span>
        {/each}
    {:else}
        {query}
    {/if}
</span>

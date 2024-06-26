<script lang="ts">
    import { decorateQuery } from '$lib/branded'
    import type { SearchPatternType } from '@sourcegraph/web/src/graphql-operations'
    import EmphasizedLabel from './EmphasizedLabel.svelte'

    export let query: string
    export let patternType: SearchPatternType | undefined = undefined
    export let matches: Set<number> | null = null

    $: decorations = decorateQuery(query, patternType)
</script>

<code class="search-query-link">
    {#if decorations}
        {#each decorations as { key, className, value, token } (key)}
            <span class={className}>
                {#if matches}
                    <EmphasizedLabel label={value} {matches} offset={token.range.start} />{:else}
                    {value}{/if}</span
            >
        {/each}
    {:else}
        {query}
    {/if}
</code>

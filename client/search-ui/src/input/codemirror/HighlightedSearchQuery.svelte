<script lang="ts">
    import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
    import { decorate, toDecoration } from '@sourcegraph/shared/src/search/query/decoratedToken'

    export let query: string

    function getDecorations(query: string) {
        const parsedQuery = scanSearchQuery(query)
        if (parsedQuery.type === 'success') {
            return parsedQuery.term.flatMap(token => decorate(token).map(token => toDecoration(query, token)))
        }
        return []
    }

    $: decorations = getDecorations(query)
</script>

<span class="text-monospace search-query-link">
    {#each decorations as decoration (decoration.key)}
        <span class={decoration.className}>{decoration.value}</span>
    {/each}
</span>

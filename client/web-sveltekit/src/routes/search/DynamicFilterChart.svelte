<script lang="ts">
    import type { SectionItem } from '$lib/search/dynamicFilters'

    export let items: SectionItem[]
</script>

{#each items as item}
    <div>{item.label}</div>
{/each}

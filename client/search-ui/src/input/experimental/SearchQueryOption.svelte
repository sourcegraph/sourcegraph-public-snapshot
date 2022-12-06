<svelte:options immutable />

<script lang="ts">
    import { Option } from './suggestions'
    import HighlightedSearchQuery from './HighlightedSearchQuery.svelte'
    export let option: Option
</script>

<HighlightedSearchQuery query={option.value} />

<script lang="ts">
    import { page } from '$app/stores'
    import { isRevisionNotFoundErrorLike } from '$lib/shared'

    import RevisionNotFoundError from './RevisionNotFoundError.svelte'
</script>

{#if isRevisionNotFoundErrorLike($page.error)}
    <RevisionNotFoundError />
{/if}
<!-- show error information for other errors -->

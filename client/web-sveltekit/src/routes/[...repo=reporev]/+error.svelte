<script lang="ts">
    import { page } from '$app/stores'
    import { isRevisionNotFoundErrorLike } from '$lib/shared'
    import { Alert } from '$lib/wildcard'

    import RevisionNotFoundError from './RevisionNotFoundError.svelte'
</script>

{#if isRevisionNotFoundErrorLike($page.error)}
    <RevisionNotFoundError />
{:else if $page.error}
    <div>
        <Alert variant="danger">
            An error occurred while loading the page: {$page.error.message}
        </Alert>
    </div>
{/if}

<style lang="scss">
    div {
        max-width: var(--viewport-xl);
        margin-left: auto;
        margin-right: auto;
        margin-top: 5rem;
    }
</style>

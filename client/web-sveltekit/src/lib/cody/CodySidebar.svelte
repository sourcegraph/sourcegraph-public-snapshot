<script lang="ts">
    import { createEventDispatcher } from 'svelte'

    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { user } from '$lib/stores'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Alert, Badge, Button } from '$lib/wildcard'

    import type { CodySidebar_ResolvedRevision } from './CodySidebar.gql'

    export let repository: CodySidebar_ResolvedRevision
    export let filePath: string

    const dispatch = createEventDispatcher<{ close: void }>()
</script>

<div class="root">
    <div class="header">
        <h3>
            <Icon icon={ISgCody} /> Cody
            <Badge variant="warning">Experimental</Badge>
        </h3>
        <Tooltip tooltip="Close Cody chat">
            <Button variant="icon" aria-label="Close Cody" on:click={() => dispatch('close')}>
                <Icon icon={ILucideX} inline aria-hidden />
            </Button>
        </Tooltip>
    </div>
    {#if $user}
        {#await import('./CodySidebarChat.svelte')}
            <LoadingSpinner />
        {:then module}
            <svelte:component this={module.default} {repository} {filePath} />
        {/await}
    {:else}
        <Alert variant="info">
            <strong>Cody is only available to signed-in users.</strong>
            <a href="/sign-in">Sign in</a> to use Cody.
        </Alert>
    {/if}
</div>

<style lang="scss">
    .root {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
    }

    .header {
        display: flex;
        flex-shrink: 0;
        align-items: center;
        justify-content: space-between;
        font-size: var(--font-size-small);
        background-color: var(--input-bg);
        border-bottom: 1px solid var(--border-color-2);
        padding: 0.25rem 1rem;

        // Shows the cody icon in color
        --icon-color: initial;

        h3 {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-weight: normal;
            margin: 0;
        }
    }
</style>

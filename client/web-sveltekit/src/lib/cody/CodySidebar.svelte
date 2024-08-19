<script context="module" lang="ts">
import { uniqueID } from "$lib/dom";

export const CODY_SIDEBAR_ID = uniqueID("cody-sidebar");
</script>

<script lang="ts">
    import { createEventDispatcher } from 'svelte'

    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { user } from '$lib/stores'
    import Tooltip from '$lib/Tooltip.svelte'
    import { Alert, Badge, Button } from '$lib/wildcard'

    import type { CodySidebar_ResolvedRevision } from './CodySidebar.gql'
    import type { LineOrPositionOrRange } from '@sourcegraph/common';

    export let repository: CodySidebar_ResolvedRevision
    export let filePath: string
    export let lineOrPosition: LineOrPositionOrRange | undefined = undefined

    const headingID = uniqueID('cody-sidebar-heading')
    const dispatch = createEventDispatcher<{ close: void }>()
</script>

<aside id={CODY_SIDEBAR_ID} aria-labelledby={headingID}>
    <div class="header">
        <div />
        <h3 id={headingID}>
            <Icon icon={ISgCody} /> Cody
            <Badge variant="info">Beta</Badge>
        </h3>
        <Tooltip tooltip="Close Cody chat">
            <Button
                variant="icon"
                aria-label="Close Cody"
                on:click={() => dispatch('close')}
                aria-controls={CODY_SIDEBAR_ID}
                aria-expanded="true"
            >
                <Icon icon={ILucideX} inline aria-hidden />
            </Button>
        </Tooltip>
    </div>
    {#if $user}
        {#await import('./CodyChat.svelte')}
            <LoadingSpinner />
        {:then module}
            <svelte:component this={module.default} {repository} {filePath} {lineOrPosition} />
        {/await}
    {:else}
        <Alert variant="info">
            <strong>Cody is only available to signed-in users.</strong>
            <a href="/sign-in">Sign in</a> to use Cody.
        </Alert>
    {/if}
</aside>

<style lang="scss">
    aside {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
        background-color: var(--body-bg);
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
        height: var(--repo-header-height);

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

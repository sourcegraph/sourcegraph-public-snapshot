<script lang="ts">
    import { sizeToFit } from '$lib/dom'
    import Icon from '$lib/Icon.svelte'
    import { pathHrefFactory } from '$lib/path'
    import ShrinkablePath from '$lib/path/ShrinkablePath.svelte'
    import { DropdownMenu } from '$lib/wildcard'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import MobileFileSidePanelOpenButton from './MobileFileSidePanelOpenButton.svelte'

    export let repoName: string
    export let revision: string | undefined
    export let path: string
    export let type: 'blob' | 'tree'

    let compact = false
    let shrinkablePath: ShrinkablePath
    function grow(): boolean {
        // Expand the breadcrumbs first, then the actions
        if (shrinkablePath.grow()) {
            return true
        }
        if (compact) {
            compact = false
            return true
        }
        return false
    }
    function shrink(): boolean {
        // Collapse the actions first, then the breadcrumbs
        if (!compact) {
            compact = true
            return true
        }
        return shrinkablePath.shrink()
    }
</script>

<header use:sizeToFit={{ grow, shrink }}>
    <h2 data-testid="file-header-path">
        <MobileFileSidePanelOpenButton />
        <ShrinkablePath
            bind:this={shrinkablePath}
            {path}
            pathHref={pathHrefFactory({ repoName, revision, fullPath: path, fullPathType: type })}
            showCopyButton
        >
            <slot name="file-icon" slot="file-icon" />
        </ShrinkablePath>
    </h2>
    <div class="actions" class:compact>
        <slot name="actions" />
        {#if $$slots.actionmenu}
            <div class="divider" />
            <div>
                <DropdownMenu
                    triggerButtonClass={getButtonClassName({ variant: 'icon' })}
                    aria-label="Show more actions"
                >
                    <svelte:fragment slot="trigger">
                        <Icon inline icon={ILucideEllipsis} aria-label="Collapsed path elements" />
                    </svelte:fragment>
                    <slot name="actionmenu" />
                </DropdownMenu>
            </div>
        {/if}
    </div>
</header>

<style lang="scss">
    header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.25rem 0 0.25rem 0.5rem;
        background-color: var(--color-bg-1);
        border-bottom: 1px solid var(--border-color);
        z-index: 1;
        gap: 1rem;
        height: var(--repo-header-height);

        @media (--mobile) {
            height: unset;
            flex-wrap: wrap;
        }
    }

    h2 {
        display: flex;
        gap: 0.5rem;
        align-items: center;
        margin: 0;

        :global([data-path-container]) {
            gap: 0.375em !important;
        }

        // Grow to fill the rest of the header so the hoverable area for
        // showing the copy button is large.
        flex: 1;

        :global([data-copy-button]) {
            opacity: 0;
            transition: opacity 0.2s;
        }
        &:is(:hover, :focus-within) :global([data-copy-button]) {
            opacity: 1;
        }
    }

    .actions {
        display: flex;
        justify-content: space-evenly;
        gap: 1rem;
        padding-right: 1rem;
        align-items: center;

        // When the actions are "compact" we hide the labels.
        &.compact {
            :global([data-action-label]) {
                display: none;
            }
        }

        .divider {
            border-left: 1px solid var(--border-color);
            align-self: stretch;

            @media (--mobile) {
                display: none;
            }
        }
    }
</style>

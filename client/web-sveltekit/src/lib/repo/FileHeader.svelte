<script lang="ts">
    import { writable } from 'svelte/store'

    import { resolveRoute } from '$app/paths'
    import { sizeToFit } from '$lib/dom'
    import Icon2 from '$lib/Icon2.svelte'
    import { DropdownMenu, MenuLink } from '$lib/wildcard'
    import { getButtonClassName } from '$lib/wildcard/Button'
    import CopyButton from '$lib/wildcard/CopyButton.svelte'

    const TREE_ROUTE_ID = '/[...repo=reporev]/(validrev)/(code)/-/tree/[...path]'
    const BLOB_ROUTE_ID = '/[...repo=reporev]/(validrev)/(code)/-/blob/[...path]'

    export let repoName: string
    export let revision: string | undefined
    export let path: string
    export let type: 'blob' | 'tree'

    $: breadcrumbs = path.split('/').map((part, index, all): [string, string] => [
        part,
        resolveRoute(
            // Only the last element in a path can be a blob
            index < all.length - 1 || type === 'tree' ? TREE_ROUTE_ID : BLOB_ROUTE_ID,
            {
                repo: revision ? `${repoName}@${revision}` : repoName,
                path: all.slice(0, index + 1).join('/'),
            }
        ),
    ])

    // HACK: we use a flexbox for the path and an inline icon, but we still want the copied path to be usable.
    // This event handler removes the newlines surrounding slashes from the copied text.
    function stripSpaces(event: ClipboardEvent) {
        const selection = document.getSelection() ?? ''
        event.clipboardData?.setData('text/plain', selection.toString().replaceAll(/\n?\/\n?/g, '/'))
        event.preventDefault()
    }

    const breadcrumbMenuOpen = writable(false)
    $: compact = false
    $: visibleBreadcrumbCount = breadcrumbs.length
    $: collapsedBreadcrumbCount = breadcrumbs.length - visibleBreadcrumbCount
    function grow(): boolean {
        // Expand the breadcrumbs first, then the actions
        if (visibleBreadcrumbCount < breadcrumbs.length) {
            visibleBreadcrumbCount += 1
            return true
        }
        compact = false
        return false
    }
    function shrink(): boolean {
        // Collapse the actions first, then the breadcrumbs
        if (!compact) {
            compact = true
            return visibleBreadcrumbCount > 1
        }
        if (visibleBreadcrumbCount > 1) {
            visibleBreadcrumbCount -= 1
        }
        return visibleBreadcrumbCount > 1
    }
</script>

<div class="header" use:sizeToFit={{ grow, shrink }}>
    <h2 on:copy={stripSpaces} data-testid="file-header-path">
        {#if collapsedBreadcrumbCount > 0}
            <DropdownMenu
                open={breadcrumbMenuOpen}
                triggerButtonClass={getButtonClassName({ variant: 'icon', outline: true, size: 'sm' })}
                aria-label="{$breadcrumbMenuOpen ? 'Close' : 'Open'} collapsed path elements"
            >
                <svelte:fragment slot="trigger">
                    <Icon2 inline icon={ILucideEllipsis} aria-label="Collapsed path elements" />
                </svelte:fragment>
                {#each breadcrumbs.slice(0, collapsedBreadcrumbCount) as [name, path]}
                    <MenuLink href={path}>
                        <Icon2 inline icon={ILucideFolder} aria-label="Collapsed path elements" />
                        {name}
                    </MenuLink>
                {/each}
            </DropdownMenu>
            <span class="slash">/</span>
        {/if}
        {#each breadcrumbs.slice(collapsedBreadcrumbCount) as [name, path], index}
            {@const last = index === breadcrumbs.length - collapsedBreadcrumbCount - 1}
            <span class:last>
                {#if index > 0}
                    <span class="slash">/</span>
                {/if}
                {#if last}
                    <slot name="icon" />
                {/if}
                {#if path}
                    <a href={path}>{name}</a>
                {:else}
                    {name}
                {/if}
            </span>
        {/each}
        <span class="copy-button"><CopyButton value={path} label="Copy path to clipboard" /></span>
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
                        <Icon2 inline icon={ILucideEllipsis} aria-label="Collapsed path elements" />
                    </svelte:fragment>
                    <slot name="actionmenu" />
                </DropdownMenu>
            </div>
        {/if}
    </div>
</div>

<style lang="scss">
    .header {
        display: flex;
        flex-wrap: nowrap;
        justify-content: space-between;
        align-items: center;
        padding: 0.25rem 0 0.25rem 0.5rem;
        background-color: var(--color-bg-1);
        border-bottom: 1px solid var(--border-color);
        z-index: 1;
        gap: 1rem;
    }

    h2 {
        flex: 1;

        display: flex;
        flex-wrap: nowrap;
        gap: 0.375em;
        span {
            display: flex;
            gap: inherit;
            white-space: nowrap;
        }

        font-weight: 400;
        font-size: var(--code-font-size);
        font-family: var(--code-font-family);
        margin: 0;

        a {
            color: var(--text-body);

            &:hover {
                color: var(--text-title);
            }
        }

        .slash {
            color: var(--text-disabled);
        }

        .last {
            color: var(--text-title);
        }

        .copy-button {
            visibility: hidden;
        }
        &:hover .copy-button {
            visibility: visible;
        }
    }

    .actions {
        --color: var(--icon-color);

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
        }
    }
</style>

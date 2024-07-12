<!--
    @component
    ShrinkablePath is a DisplayPath that can collapse its path items
    into a dropdown menu to save space. It does not do this automatically,
    and is usually used alongside a helper like `sizeToFit`.
-->
<script lang="ts">
    import { writable } from 'svelte/store'

    import { getButtonClassName } from '$lib/wildcard/Button'
    import DropdownMenu from '$lib/wildcard/menu/DropdownMenu.svelte'
    import MenuLink from '$lib/wildcard/menu/MenuLink.svelte'
    import MenuText from '$lib/wildcard/menu/MenuText.svelte'

    import Icon from '../Icon.svelte'

    import DisplayPath from './DisplayPath.svelte'

    export let path: string
    export let showCopyButton = false
    export let pathHref: ((path: string) => string) | undefined = undefined

    $: parts = path.split('/').map((part, index, allParts) => ({ part, path: allParts.slice(0, index + 1).join('/') }))
    let collapsedPartCount = 0
    $: collapsedParts = parts.slice(0, collapsedPartCount)
    $: visibleParts = parts.slice(collapsedPartCount)

    $: scopedPathHref = pathHref
        ? (path: string) => pathHref(collapsedParts.map(({ part }) => part).join('/') + path)
        : undefined

    // Returns whether shrinking was successful
    export function shrink(): boolean {
        // Never collapse the last element of the path
        if (collapsedPartCount < parts.length - 1) {
            collapsedPartCount++
            return true
        }
        return false
    }

    // Returns whether growing was successful
    export function grow(): boolean {
        if (collapsedPartCount > 0) {
            collapsedPartCount--
            return true
        }
        return false
    }

    const breadcrumbMenuOpen = writable(false)
</script>

<DisplayPath path={visibleParts.map(({ part }) => part).join('/')} {showCopyButton} pathHref={scopedPathHref}>
    <svelte:fragment slot="prefix">
        {#if collapsedParts.length > 0}
            <DropdownMenu
                open={breadcrumbMenuOpen}
                triggerButtonClass={getButtonClassName({ variant: 'icon', outline: true, size: 'sm' })}
                aria-label="{$breadcrumbMenuOpen ? 'Close' : 'Open'} collapsed path elements"
            >
                <Icon slot="trigger" inline icon={ILucideEllipsis} aria-label="Collapsed path elements" />
                {#each collapsedParts as { part, path }}
                    <svelte:component
                        this={pathHref ? MenuLink : MenuText}
                        href={pathHref ? pathHref(path) : undefined}
                    >
                        <Icon inline icon={ILucideFolder} aria-hidden="true" />
                        {part}
                    </svelte:component>
                {/each}
            </DropdownMenu>
            <span data-slash>/</span>
        {/if}
    </svelte:fragment>
    <slot name="file-icon" slot="file-icon" />
</DisplayPath>

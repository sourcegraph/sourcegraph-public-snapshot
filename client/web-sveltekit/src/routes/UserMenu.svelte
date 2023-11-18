<script lang="ts">
    import Icon from '$lib/Icon.svelte'
    import UserAvatar from '$lib/UserAvatar.svelte'
    import type { AuthenticatedUser } from '$lib/shared'
    import { humanTheme } from '$lib/theme'
    import { DropdownMenu, MenuLink, MenuRadioGroup, MenuSeparator, Submenu } from '$lib/wildcard'
    import { mdiChevronDown, mdiChevronUp, mdiOpenInNew } from '@mdi/js'
    import { writable } from 'svelte/store'

    const MAX_VISIBLE_ORGS = 5

    export let authenticatedUser: AuthenticatedUser

    const open = writable(false)
    $: organizations = authenticatedUser.organizations.nodes
</script>

<DropdownMenu {open} aria-label="{$open ? 'Close' : 'Open'} user profile menu">
    <svelte:fragment slot="trigger">
        <UserAvatar user={authenticatedUser} />
        <Icon svgPath={$open ? mdiChevronUp : mdiChevronDown} aria-hidden={true} inline />
    </svelte:fragment>
    <h6>Signed in as <strong>@{authenticatedUser.username}</strong></h6>
    <MenuSeparator />
    <MenuLink href={authenticatedUser.settingsURL} data-sveltekit-reload>Settings</MenuLink>
    <MenuLink href="/users/{authenticatedUser.username}/searches" data-sveltekit-reload>Saved searches</MenuLink>
    <MenuLink href="/teams" data-sveltekit-reload>Teams</MenuLink>
    <MenuSeparator />
    <Submenu>
        <svelte:fragment slot="trigger">Theme</svelte:fragment>
        <MenuRadioGroup values={['Light', 'Dark', 'System']} value={humanTheme} />
    </Submenu>
    {#if organizations.length > 0}
        <MenuSeparator />
        <h6>Your organizations</h6>
        {#each organizations.slice(0, MAX_VISIBLE_ORGS) as org}
            <MenuLink href={org.settingsURL || org.url}>
                {org.displayName || org.name}
            </MenuLink>
        {/each}
        {#if organizations.length > MAX_VISIBLE_ORGS}
            <MenuLink href={authenticatedUser.settingsURL}>Show all organizations</MenuLink>
        {/if}
    {/if}
    <MenuSeparator />
    {#if authenticatedUser.siteAdmin}
        <MenuLink href="/site-admin" data-sveltekit-reload>Site admin</MenuLink>
    {/if}
    <MenuLink href="/help" target="_blank" rel="noopener">
        Help <Icon aria-hidden={true} svgPath={mdiOpenInNew} inline />
    </MenuLink>
    <MenuLink href="/-/sign-out" data-sveltekit-reload>Sign out</MenuLink>
</DropdownMenu>

<style lang="scss">
    h6 {
        padding: var(--dropdown-item-padding);
        margin: 0;
        font-size: 0.75rem;
        color: var(--dropdown-header-color);
    }
</style>

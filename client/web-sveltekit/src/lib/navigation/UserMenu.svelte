<script lang="ts">
    import { writable } from 'svelte/store'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import { ThemeSetting, themeSetting } from '$lib/theme'
    import { DropdownMenu, MenuLink, MenuRadioGroup, MenuSeparator, Submenu } from '$lib/wildcard'
    import { getButtonClassName } from '$lib/wildcard/Button'

    import type { UserMenu_User } from './UserMenu.gql'

    const MAX_VISIBLE_ORGS = 5

    export let user: UserMenu_User

    const open = writable(false)
    $: organizations = user.organizations.nodes
    $: ownEnabled = window.context.ownEnabled
    $: isDotCom = window.context.sourcegraphDotComMode
</script>

<DropdownMenu
    {open}
    triggerButtonClass={getButtonClassName({ variant: 'icon' })}
    aria-label="{$open ? 'Close' : 'Open'} user menu"
>
    <svelte:fragment slot="trigger">
        <Avatar avatar={user} --avatar-size="1.5rem" />
        <Icon icon={$open ? ILucideChevronUp : ILucideChevronDown} aria-hidden={true} inline />
    </svelte:fragment>
    <h6>Signed in as <strong>@{user.username}</strong></h6>
    <MenuSeparator />
    <MenuLink href={user.settingsURL}>Settings</MenuLink>
    <MenuLink href="/users/{user.username}/searches">Saved searches</MenuLink>
    <!--  TODO add a check for dotcom and for whether own is enabled -->
    {#if ownEnabled && !isDotCom}
        <MenuLink href="/teams">Teams</MenuLink>
    {/if}
    <MenuSeparator />
    <Submenu>
        <svelte:fragment slot="trigger">Theme</svelte:fragment>
        <MenuRadioGroup values={[ThemeSetting.Light, ThemeSetting.Dark, ThemeSetting.System]} value={themeSetting} />
    </Submenu>
    {#if organizations.length > 0}
        <MenuSeparator />
        <h6>Your organizations</h6>
        {#each organizations.slice(0, MAX_VISIBLE_ORGS) as org}
            <MenuLink href={org.settingsURL || org.url}>
                {org.name}
            </MenuLink>
        {/each}
        {#if organizations.length > MAX_VISIBLE_ORGS}
            <MenuLink href={user.settingsURL}>Show all organizations</MenuLink>
        {/if}
    {/if}
    <MenuSeparator />
    {#if user.siteAdmin}
        <MenuLink href="/site-admin">Site admin</MenuLink>
    {/if}
    <MenuLink href="/help" target="_blank" rel="noopener">
        Help <Icon aria-hidden={true} icon={ILucideExternalLink} inline />
    </MenuLink>
    <MenuLink href="/-/sign-out">Sign out</MenuLink>
</DropdownMenu>

<style lang="scss">
    h6 {
        padding: var(--dropdown-item-padding);
        margin: 0;
        font-size: 0.75rem;
        color: var(--dropdown-header-color);
    }
</style>

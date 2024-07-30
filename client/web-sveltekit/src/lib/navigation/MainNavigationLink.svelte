<script lang="ts">
    import type { HTMLAnchorAttributes } from 'svelte/elements'

    import Icon from '$lib/Icon.svelte'
    import { Badge } from '$lib/wildcard'

    import { type NavigationEntry, Status } from './mainNavigation'

    type $$Props = {
        entry: NavigationEntry
    } & HTMLAnchorAttributes

    export let entry: NavigationEntry
</script>

<a href={entry.href} {...$$restProps}>
    {#if entry.icon}
        <Icon icon={entry.icon} aria-hidden="true" inline />&nbsp;
    {/if}
    {entry.label}
    {#if entry.status && entry.status & Status.BETA}
        &nbsp;
        <Badge variant="info">Beta</Badge>
    {/if}
</a>

<style lang="scss">
    a {
        display: inline-flex;
        align-items: center;
        text-decoration: none;
        color: var(--body-color);

        &:hover {
            color: inherit;
            text-decoration: none;
        }
    }
</style>

<script lang="ts">
    import type { HTMLAnchorAttributes } from 'svelte/elements'
    import { getContext } from './DropdownMenu.svelte'

    type $$Props = {
        href: HTMLAnchorAttributes['href']
    } & HTMLAnchorAttributes

    const { item } = getContext()

    export let href: HTMLAnchorAttributes['href']
</script>

<a {href} {...$$restProps} {...$item} use:item>
    <slot />
</a>

<!-- styles are defined in Menu.svelte -->

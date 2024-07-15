<script lang="ts">
    import type { HTMLAttributes } from 'svelte/elements'

    import { getContext } from './DropdownMenu.svelte'

    type $$Props = HTMLAttributes<HTMLDivElement>

    const { item } = getContext()
</script>

<div {...$$restProps} {...$item} use:item>
    <slot />
</div>

<!-- styles are mostly defined in Menu.svelte. Unsets default styles for non-interactive menu item -->
<style lang="scss">
    div[role='menuitem'] {
        cursor: unset;
        &:hover,
        &:focus {
            background-color: unset;
            color: unset;
            text-decoration: none;
        }
    }
</style>

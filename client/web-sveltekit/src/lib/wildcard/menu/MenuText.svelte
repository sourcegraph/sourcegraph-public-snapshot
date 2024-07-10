<script lang="ts">
    import type { HTMLAttributes } from 'svelte/elements'

    import { getContext } from './DropdownMenu.svelte'

    type $$Props = HTMLAttributes<HTMLDivElement>

    const { item } = getContext()
</script>

<div {...$$restProps} {...$item} use:item>
    <slot />
</div>

<!-- styles are defined in Menu.svelte -->

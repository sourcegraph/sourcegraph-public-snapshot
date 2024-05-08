<script lang="ts">
    import type { HTMLButtonAttributes } from 'svelte/elements'
    import { getContext } from './DropdownMenu.svelte'

    type $$Props = HTMLButtonAttributes

    const { item } = getContext()
</script>

<button {...$$restProps} {...$item} use:item on:click>
    <slot />
</button>

<!-- styles are defined in Menu.svelte -->

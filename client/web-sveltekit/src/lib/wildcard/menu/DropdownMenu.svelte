<script lang="ts" context="module">
    import { createContextAccessors } from '$lib/utils/context'

    type DropdownMenu = ReturnType<typeof createDropdownMenu>

    interface DropdownMenuContext {
        item: DropdownMenu['elements']['item']
        separator: DropdownMenu['elements']['separator']
        builders: DropdownMenu['builders']
    }

    const [setContext, getContext] = createContextAccessors<DropdownMenuContext>()

    export { getContext }
</script>

<script lang="ts">
    import { createDropdownMenu } from '@melt-ui/svelte'
    import type { HTMLButtonAttributes } from 'svelte/elements'
    import { writable, type Writable } from 'svelte/store'

    type $$Props = {
        open?: Writable<boolean>
        triggerButtonClass: string
    } & HTMLButtonAttributes

    export let triggerButtonClass: string
    export let open: Writable<boolean> = writable(false)

    const {
        elements: { menu, item, trigger, separator },
        builders,
    } = createDropdownMenu({
        open,
    })
    setContext({ item, separator, builders })
</script>

<button data-dropdown-trigger {...$trigger} use:trigger class={triggerButtonClass} {...$$restProps}>
    <slot name="trigger" />
</button>

<div {...$menu} use:menu>
    <slot />
</div>

<style lang="scss">
    div,
    div :global([role='menu']) {
        isolation: isolate;
        min-width: 12rem;
        font-size: var(--font-size-small);
        background-clip: padding-box;
        background-color: var(--dropdown-bg);
        border: 1px solid var(--dropdown-border-color);
        border-radius: var(--popover-border-radius);
        color: var(--body-color);
        box-shadow: var(--dropdown-shadow);
        padding: var(--dropdown-padding-y) 0;

        :global([role^='menuitem']) {
            --icon-color: currentColor;

            all: unset;
            cursor: pointer;
            display: block;
            box-sizing: border-box;
            width: 100%;
            padding: var(--dropdown-item-padding);
            white-space: nowrap;
            color: var(--dropdown-link-hover-color);

            &:disabled {
                color: var(--text-muted);
                cursor: not-allowed;
            }

            &:hover,
            &:focus {
                background-color: var(--dropdown-link-hover-bg);
                color: var(--dropdown-link-hover-color);
                text-decoration: none;
            }
        }
    }
</style>

<script lang="ts">
    import type { HTMLAttributes } from 'svelte/elements'
    import type { Writable } from 'svelte/store'

    import { getContext } from './DropdownMenu.svelte'

    type $$Props = {
        values: string[]
        value: Writable<string>
    } & HTMLAttributes<HTMLDivElement>

    export let values: string[]
    export let value: Writable<string>

    const {
        elements: { radioItem, radioGroup },
        helpers: { isChecked },
    } = getContext().builders.createMenuRadioGroup({
        value,
    })
</script>

<div {...$radioGroup} {...$$restProps} use:radioGroup>
    {#each values as value}
        {@const checked = $isChecked(value)}
        <div {...$radioItem({ value })} use:radioItem>
            <input type="radio" {checked} aria-hidden="true" /><!--
            --><span>
                <slot {value} {checked}>
                    {value}
                </slot>
            </span>
        </div>
    {/each}
</div>

<style lang="scss">
    span {
        margin-left: 0.5rem;
    }
</style>

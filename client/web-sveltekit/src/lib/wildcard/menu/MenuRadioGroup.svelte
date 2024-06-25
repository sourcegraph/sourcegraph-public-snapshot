<script lang="ts">
    import type { Writable } from 'svelte/store'

    import { getContext } from './DropdownMenu.svelte'

    export let values: string[]
    export let value: Writable<string>

    const {
        elements: { radioItem, radioGroup },
        helpers: { isChecked },
    } = getContext().builders.createMenuRadioGroup({
        value,
    })
</script>

<div {...$radioGroup} use:radioGroup>
    {#each values as value}
        {@const checked = $isChecked(value)}
        <div class="item" {...$radioItem({ value })} use:radioItem>
            <input type="radio" {value} {checked} />&nbsp
            <slot {value} {checked}>
                {value}
            </slot>
        </div>
    {/each}
</div>

<style lang="scss">
    span {
        float: right;
    }
</style>

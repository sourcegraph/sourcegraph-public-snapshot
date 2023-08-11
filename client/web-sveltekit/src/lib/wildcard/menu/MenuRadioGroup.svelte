<script lang="ts">
    import type { Writable } from 'svelte/store'
    import { getContext } from './Menu.svelte'
    import { mdiCheck } from '@mdi/js'
    import Icon from '$lib/Icon.svelte'

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
            <slot {value} {checked}>
                {value}
            </slot>
            {#if checked}
                <span>
                    <Icon svgPath={mdiCheck} inline />
                </span>
            {/if}
        </div>
    {/each}
</div>

<style lang="scss">
    span {
        float: right;
    }
</style>

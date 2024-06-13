<script lang="ts">
    import { mdiChevronRight } from '@mdi/js'
    import { getContext } from './DropdownMenu.svelte'
    import Icon2 from '$lib/Icon2.svelte'

    const {
        elements: { subMenu, subTrigger },
    } = getContext().builders.createSubmenu()
</script>

<div {...$subTrigger} use:subTrigger>
    <slot name="trigger" />
    <span>
        <Icon2 icon={ILucideChevronRight} inline aria-hidden />
    </span>
</div>
<div {...$subMenu} use:subMenu>
    <slot />
</div>

<style lang="scss">
    span {
        float: right;
    }
</style>

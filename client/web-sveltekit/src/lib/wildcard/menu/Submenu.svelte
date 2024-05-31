<script lang="ts">
    import Icon2 from '$lib/Icon2.svelte'

    import { getContext } from './DropdownMenu.svelte'

    const {
        elements: { subMenu, subTrigger },
    } = getContext().builders.createSubmenu()
</script>

<div {...$subTrigger} use:subTrigger>
    <slot name="trigger" />
    <span>
        <Icon2 icon={ILucideChevronRight} inline />
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

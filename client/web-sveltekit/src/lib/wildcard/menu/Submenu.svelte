<script lang="ts">
    import Icon from '$lib/Icon.svelte'

    import { getContext } from './DropdownMenu.svelte'

    const {
        elements: { subMenu, subTrigger },
    } = getContext().builders.createSubmenu()
</script>

<div {...$subTrigger} use:subTrigger>
    <slot name="trigger" />
    <span>
        <Icon icon={ILucideChevronRight} inline aria-hidden />
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

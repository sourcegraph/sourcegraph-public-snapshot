<script lang="ts">
    import { mdiChevronRight } from '@mdi/js'
    import { getContext } from './Menu.svelte'
    import Icon from '$lib/Icon.svelte'

    const {
        elements: { subMenu, subTrigger },
    } = getContext().builders.createSubmenu()
</script>

<div {...$subTrigger} use:subTrigger>
    <slot name="trigger" />
    <span>
        <Icon svgPath={mdiChevronRight} inline />
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

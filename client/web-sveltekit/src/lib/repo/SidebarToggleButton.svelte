<script lang="ts">
    import { mdiChevronDoubleLeft, mdiChevronDoubleRight } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import Button from '$lib/wildcard/Button.svelte'

    import { sidebarOpen } from './stores'

    $: tooltip = ($sidebarOpen ? 'Hide' : 'Show') + ' sidebar'
    $: svgPath = $sidebarOpen ? mdiChevronDoubleLeft : mdiChevronDoubleRight
</script>

<Tooltip {tooltip}>
    <Button variant="secondary" outline size="sm" on:click={() => sidebarOpen.update(open => !open)}>
        <Icon {svgPath} inline />
    </Button>
</Tooltip>

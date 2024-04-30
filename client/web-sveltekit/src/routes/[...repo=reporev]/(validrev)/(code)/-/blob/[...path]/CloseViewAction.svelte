<script lang="ts">
    import { mdiClose } from '@mdi/js'

    import { page } from '$app/stores'
    import { SourcegraphURL } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    export let label: string

    $: href = SourcegraphURL.from($page.url).deleteSearchParameter('rev', 'diff').toString()
</script>

<Tooltip tooltip={label}>
    <a {href}><Icon svgPath={mdiClose} inline /> <span data-action-label>{label}</span></a>
</Tooltip>

<style lang="scss">
    a,
    a:hover {
        color: var(--text-color);
        text-decoration: none;
        white-space: nowrap;
    }
</style>

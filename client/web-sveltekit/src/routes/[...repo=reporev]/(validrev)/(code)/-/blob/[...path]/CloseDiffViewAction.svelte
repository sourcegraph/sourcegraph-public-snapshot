<script lang="ts">
    import { mdiClose } from '@mdi/js'

    import { page } from '$app/stores'
    import { SourcegraphURL } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'

    $: href = SourcegraphURL.from($page.url).deleteSearchParameter('rev').toString()
</script>

<Tooltip tooltip="Close diff">
    <a {href}><Icon svgPath={mdiClose} inline /> <span data-action-label>Close diff</span></a>
</Tooltip>

<style lang="scss">
    a,
    a:hover {
        color: var(--text-color);
        text-decoration: none;
        white-space: nowrap;
    }
</style>

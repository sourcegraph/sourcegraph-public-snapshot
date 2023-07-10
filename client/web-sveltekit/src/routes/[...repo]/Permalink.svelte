<script lang="ts">
    import { mdiLink } from '@mdi/js'

    import { page } from '$app/stores'
    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { replaceRevisionInURL } from '$lib/web'

    $: resolvedRevision = isErrorLike($page.data.resolvedRevision) ? null : $page.data.resolvedRevision

    $: href = resolvedRevision
        ? replaceRevisionInURL($page.url.pathname + $page.url.search + $page.url.hash, resolvedRevision.commitID)
        : ''
</script>

{#if href}
    <Tooltip tooltip="Permalink (with full Git commit SHA)">
        <a {href}><Icon svgPath={mdiLink} inline /></a>
    </Tooltip>
{/if}

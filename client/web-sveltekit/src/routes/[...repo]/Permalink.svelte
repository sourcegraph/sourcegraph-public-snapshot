<script lang="ts">
    import { mdiLink } from '@mdi/js'

    import { page } from '$app/stores'
    import { replaceRevisionInURL } from '$lib/web'
    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'

    $: resolvedRevision = isErrorLike($page.data.resolvedRevision) ? null : $page.data.resolvedRevision

    $: href = resolvedRevision
        ? replaceRevisionInURL($page.url.pathname + $page.url.search + $page.url.hash, resolvedRevision.commitID)
        : ''
</script>

{#if href}
    <a {href}><Icon svgPath={mdiLink} inline /></a>
{/if}

<script lang="ts">
    import { mdiLink } from '@mdi/js'

    import { page } from '$app/stores'
    import { isErrorLike, type ErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { replaceRevisionInURL, type RepoResolvedRevision } from '$lib/web'

    export let resolvedRevision: RepoResolvedRevision | ErrorLike

    $: href = !isErrorLike(resolvedRevision)
        ? replaceRevisionInURL($page.url.pathname + $page.url.search + $page.url.hash, resolvedRevision.commitID)
        : ''
</script>

{#if href}
    <Tooltip tooltip="Permalink (with full Git commit SHA)">
        <a {href}><Icon svgPath={mdiLink} inline /></a>
    </Tooltip>
{/if}

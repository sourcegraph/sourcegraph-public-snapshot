<script lang="ts">
    import { mdiLink } from '@mdi/js'

    import { page } from '$app/stores'
    import { isErrorLike, type ErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { replaceRevisionInURL } from '$lib/web'
    import type { ResolvedRevision } from '$lib/repo/api/repo'

    export let resolvedRevision: ResolvedRevision | ErrorLike

    $: href = !isErrorLike(resolvedRevision)
        ? replaceRevisionInURL($page.url.toString(), resolvedRevision.commitID)
        : ''
</script>

{#if href}
    <Tooltip tooltip="Permalink (with full Git commit SHA)">
        <a {href}><Icon svgPath={mdiLink} inline /></a>
    </Tooltip>
{/if}

<script lang="ts">
    import { mdiLink } from '@mdi/js'

    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'
    import Tooltip from '$lib/Tooltip.svelte'
    import { replaceRevisionInURL } from '$lib/web'

    export let commitID: string

    $: href = commitID ? replaceRevisionInURL($page.url.toString(), commitID) : ''
</script>

{#if href}
    <Tooltip tooltip="Permalink (with full Git commit SHA)">
        <a {href}><Icon svgPath={mdiLink} inline /></a>
    </Tooltip>
{/if}

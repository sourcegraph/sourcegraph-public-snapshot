<!--
    @component
    Renders a permalink to the current page with the given Git commit ID.
-->
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
    <Tooltip tooltip="Permalink (with full git commit SHA)">
        <a {href}><Icon svgPath={mdiLink} inline /> <span data-action-label>Permalink</span></a>
    </Tooltip>
{/if}

<style lang="scss">
    a {
        color: var(--body-color);
        text-decoration: none;
        white-space: nowrap;
    }
</style>

<!--
    @component
    Renders a permalink to the current page with the given Git commit ID.
-->
<script lang="ts">
    import { mdiLink } from '@mdi/js'

    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'
    import { replaceRevisionInURL } from '$lib/shared'
    import Tooltip from '$lib/Tooltip.svelte'

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
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.5rem;
        color: var(--text-body);
        text-decoration: none;
        white-space: nowrap;

        &:hover {
            color: var(--text-title);
        }
    }
</style>

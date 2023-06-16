<script lang="ts">
    import { mdiEye } from '@mdi/js'

    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'

    $: formatted = $page.url.searchParams.get('view') !== 'raw'
    let url: string
    $: if (formatted) {
        const newURL = new URL($page.url)
        newURL.searchParams.set('view', 'raw')
        url = newURL.toString()
    } else {
        const newURL = new URL($page.url)
        newURL.searchParams.delete('view')
        url = newURL.toString()
    }
</script>

<button on:click={() => goto(url)}><Icon svgPath={mdiEye} inline /> {formatted ? 'Raw' : 'Formatted'}</button>

<style lang="scss">
    button {
        margin: 0;
        padding: 0;
        border: 0;
        background-color: transparent;
        cursor: pointer;
        margin-right: 1rem;
    }
</style>

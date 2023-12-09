<script lang="ts">
    import { page } from '$app/stores'
    import Icon from '$lib/Icon.svelte'
    import { mdiInformation, mdiReact } from '@mdi/js'

    // A hard referesh will only load the React version if the URL does not explicitly
    // include the `feat=enable-sveltekit` query parameter.
    $: urlHasFeatureFlag = $page.url.searchParams.getAll('feat').includes('enable-sveltekit')
</script>

<div>
    <Icon svgPath={mdiInformation} inline />
    You are trying out the SvelteKit prototype. Links marked with <Icon svgPath={mdiReact} inline /> will bring you back
    to the React version{#if urlHasFeatureFlag}.{:else}, as will a hard refresh.{/if}
</div>

<style lang="scss">
    div {
        padding: 0.5rem;
        background: var(--primary-4);
        --color: var(--primary-3);
    }
</style>

<script lang="ts">
    // @sg EnableRollout
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { Alert, Button } from '$lib/wildcard'
    import ProductStatusBadge from '$lib/wildcard/ProductStatusBadge.svelte'

    import type { PageData } from './$types'

    export let data: PageData
</script>

<svelte:head>
    <title>Cody Chat - Sourcegraph</title>
</svelte:head>

<section>
    <header>
        <h2>
            <Icon icon={ISgCody} aria-hidden --icon-color="initial" />
            <span>Cody Chat</span>
            <ProductStatusBadge status="beta" />
        </h2>

        <div class="actions">
            <a href={data.dashboardRoute}>Editor extensions</a>
            &nbsp;
            <Button variant="secondary">
                <a slot="custom" let:buttonClass class={buttonClass} href={data.dashboardRoute}> Dashboard </a>
            </Button>
        </div>
    </header>

    {#await import('$lib/cody/CodyChat.svelte')}
        <LoadingSpinner center />
    {:then module}
        <svelte:component this={module.default} />
    {:catch}
        <Alert variant="warning">Failed to load Cody Chat</Alert>
    {/await}
</section>

<style lang="scss">
    section {
        max-width: var(--viewport-lg);
        width: 100%;
        margin: 2rem auto 0 auto;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }

    header {
        display: flex;
        justify-content: space-between;

        margin-bottom: 1rem;
    }

    h2 > * {
        vertical-align: middle;
    }
</style>

<script lang="ts">
    import { page } from '$app/stores'
    import TabsHeader from '$lib/TabsHeader.svelte'

    import type { LayoutData } from './$types'

    export let data: LayoutData

    const tabs = [
        { id: 'overview', title: 'Overview', href: `${data.repoURL}/-/branches` },
        { id: 'all', title: 'All branches', href: `${data.repoURL}/-/branches/all` },
    ]
    $: selected = tabs.findIndex(tab => tab.href === $page.url.pathname)
</script>

<section>
    <div class="header">
        <TabsHeader {tabs} id="branche page" {selected} />
    </div>
    <slot />
</section>

<style lang="scss">
    section {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
        gap: 1rem;
        padding: 0.5rem 0;

        @media (--mobile) {
            padding: 0;
        }
    }

    .header {
        --tabs-header-align: center;

        &::after {
            content: '';
            display: block;
            border-bottom: 1px solid var(--border-color);
            position: absolute;
            left: 0;
            right: 0;
        }
    }
</style>

<script lang="ts">
    import { page } from '$app/stores'

    import type { LayoutData } from './$types'

    export let data: LayoutData

    $: subpage = $page.url.pathname.includes('/branches/all') ? 'all' : 'overview'
</script>

<section>
    <ul>
        <li class:active={subpage === 'overview'}><a href="{data.repoURL}/-/branches">Overview</a></li>
        <li class:active={subpage === 'all'}><a href="{data.repoURL}/-/branches/all">All branches</a></li>
    </ul>
    <slot />
</section>

<style lang="scss">
    section {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;
    }

    ul {
        align-self: center;
        max-width: var(--viewport-xl);
        width: 100%;

        padding: 1rem;
        margin: 0;

        display: flex;
        gap: 1rem;

        list-style: none;

        li {
            &.active {
                font-weight: bold;
            }
        }
    }
</style>

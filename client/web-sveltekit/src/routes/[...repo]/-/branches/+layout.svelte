<script lang="ts">
    import { page } from '$app/stores'

    import type { LayoutData } from './$types'

    export let data: LayoutData

    $: subpage = $page.url.pathname.includes('/branches/all') ? 'all' : 'overview'
</script>

<section>
    <div>
        <ul class="mb-3">
            <li class:active={subpage === 'overview'}><a href="{data.repoURL}/-/branches">Overview</a></li>
            <li class:active={subpage === 'all'}><a href="{data.repoURL}/-/branches/all">All branches</a></li>
        </ul>
        <slot />
    </div>
</section>

<style lang="scss">
    ul {
        margin: 0;
        padding: 0;
        list-style: none;
        display: flex;

        li {
            margin-right: 1rem;

            &.active {
                font-weight: bold;
            }
        }
    }

    section {
        overflow: auto;
        margin-top: 2rem;
    }

    div {
        max-width: 54rem;
        margin-left: auto;
        margin-right: auto;
    }
</style>

<script lang="ts">
    import Commit from '$lib/Commit.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    const { pending, value: commits, set } = createPromiseStore<PageData['deferred']['commits']>()
    $: set(data.deferred.commits)
</script>

<section>
    <div>
        <h2>View commits from this repsitory</h2>
        <h3>Changes</h3>
        {#if $pending}
            Loading...
        {:else if $commits}
            <ul>
                {#each $commits as commit (commit.canonicalURL)}
                    <li><Commit {commit} /></li>
                {/each}
            </ul>
        {/if}
    </div>
</section>

<style lang="scss">
    ul {
        list-style: none;
        padding: 0;
        margin: 0;
        flex: 1;
        width: 100%;
    }

    section {
        overflow: auto;
        margin-top: 1rem;
    }

    div {
        max-width: 54rem;
        margin-left: auto;
        margin-right: auto;
    }

    li {
        border-bottom: 1px solid var(--border-color);
        padding: 0.5rem 0;

        &:last-child {
            border: none;
        }
    }
</style>

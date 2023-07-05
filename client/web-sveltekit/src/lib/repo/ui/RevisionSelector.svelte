<script lang="ts">
    import Icon from "$lib/Icon.svelte"
    import Popover from "$lib/Popover.svelte"
    import { GitRefType } from "$lib/graphql-operations"
    import { queryGitReferences } from "$lib/loader/repo"
    import { getRelativeTime } from "$lib/relativeTime"
    import { currentDate } from "$lib/stores"
import type { ResolvedRevision } from "$lib/web"
    import Button from "$lib/wildcard/Button.svelte"
    import { mdiSourceBranch, mdiTriangleSmallDown } from "@mdi/js"
    import { getRevisionLabel } from "../utils"
    import Shimmer from "$lib/Shimmer.svelte"

    /**
     * Symbolic revesion or hash usually extracted from the URL.
     */
    export let repoID: string
    export let revision: string
    export let resolvedRevision: ResolvedRevision|null
    export let createURL: (revision: string) => string

    $: label = getRevisionLabel(revision, resolvedRevision)
    let query = ''
    $: branches = queryGitReferences({
            repo: repoID,
            type: GitRefType.GIT_BRANCH,
            first: 10,
            query,
        }).toPromise().then(result => result.nodes.map(node => ({
                        revision: node.abbrevName,
                    name: node.displayName ?? node.abbrevName ?? node.name,
                    date: node.target.commit ? new Date(node.target.commit.author.date) : undefined
        }))
        )
</script>


<Popover let:toggle let:isOpen let:registerTrigger placement="bottom-start">
    <Button variant="secondary" size="sm">
        <button
            slot="custom"
            let:className
            class="{className}"
            type="button"
            aria-expanded={isOpen}
            on:click={() => toggle()}
            use:registerTrigger
        >
        <Icon svgPath={mdiSourceBranch} inline/> {label} <Icon svgPath={mdiTriangleSmallDown} inline/>
        </button>
    </Button>
    <div slot="content" class="content" let:toggle>
        <h4>Branches</h4>
        <ul>
        {#await branches}
                <li style="width: 15rem"><Shimmer /></li>
        {:then branches}
            {#each branches as branch}
                <li>
                    <a href={createURL(branch.revision)} on:click={() => toggle(false)}>
                            <span class="name"><Icon svgPath={mdiSourceBranch} inline/>{branch.name}</span>
                            <span class="timestamp text-muted">
                                {#if branch.date}
                                    {getRelativeTime(branch.date, $currentDate)}
                                {/if}
                            </span>
                    </a>
                </li>
            {/each}
        {/await}
        </ul>
    </div>
</Popover>


<style lang="scss">
    .content {
        padding: 1rem;
        max-width: 25rem;
    }

    a {
        display: flex;
        justify-content: space-between;

        .name {
            overflow: hidden;
            text-overflow: ellipsis;
            flex: 1 1 0;
        }
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;

        li {
            padding: 0.2rem 0;

        }
    }
</style>

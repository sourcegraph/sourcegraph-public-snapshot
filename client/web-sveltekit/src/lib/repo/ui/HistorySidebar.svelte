<script lang="ts">
    import Icon from "$lib/Icon.svelte"
import LoadingSpinner from "$lib/LoadingSpinner.svelte"
    import { mdiSourceCommit } from "@mdi/js"
    import { fetchHistory } from "../api/history"
    import { getRelativeTime } from "$lib/relativeTime"
    import {currentDate} from '$lib/stores'
    import { page } from "$app/stores"

    export let repoID: string
    export let revision: string
    export let path: string

    const INITIAL_PAGE_SIZE = 10

    let first = INITIAL_PAGE_SIZE

    $: commits = fetchHistory(repoID, revision, path, first)
    $: currentRev = $page.url.searchParams.get('rev')

    function getCompareURL(rev: string): string {
        const url = new URL($page.url)
        url.searchParams.set('rev', rev)
        return url.toString()
    }
</script>


{#await commits}
    <LoadingSpinner />
{:then commits}
    <ul>
        {#each commits as commit}
            <li class:selected={currentRev === commit.oid}>
                <span class="badge">
                    <Icon svgPath={mdiSourceCommit}/>
                </span>
                <a href={getCompareURL(commit.oid)}>
                    {commit.subject}<br/>
                    <span class="text-muted">{getRelativeTime(new Date(commit.author.date), $currentDate)}</span>
                </a>
            </li>
        {/each}
    </ul>
{/await}


<style lang="scss">
    ul {
        margin: 0 0.5rem;
        padding: 0;
        list-style: none;
        margin-top: 1rem;

        li {
            position: relative;
            display: flex;
            padding: 0.2rem 0;
            border-radius: var(--border-radius);
            padding: 0 1rem;

            &:hover, &.selected {
                background-color: var(--color-bg-1);

                .badge {
                    background-color: inherit;
                }
            }

            a {
                text-decoration: none;
            }

            .badge {
                margin-left: -11px;
                z-index: 1;
                background-color: var(--body-bg);
                align-self: start;
            }

            &::before {
                content: " ";
                position: absolute;
                top: 0;
                bottom: 0;
                left: 1rem;
                width: 2px;
                background-color: black;
            }
        }
    }
</style>

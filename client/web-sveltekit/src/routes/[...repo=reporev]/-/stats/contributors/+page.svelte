<script lang="ts">
    import { pluralize } from '@sourcegraph/common'

    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import Avatar from '$lib/Avatar.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Paginator from '$lib/Paginator.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { createPromiseStore } from '$lib/utils'
    import { Alert, Button, ButtonGroup } from '$lib/wildcard'

    import type { PageData } from './$types'
    import type { ContributorConnection } from './page.gql'

    export let data: PageData

    const timePeriodButtons = [
        ['Last 7 days', '7 days ago'],
        ['Last 30 days', '30 days ago'],
        ['Last year', '1 year ago'],
        ['All time', ''],
    ]

    const contributorConnection = createPromiseStore<ContributorConnection | null>()
    $: contributorConnection.set(data.contributors)

    // We want to show stale contributors data when the user navigates to
    // the next or previous page for the current time period. When the user
    // changes the time period we want to show a loading indicator instead.
    let currentContributorConnection = $contributorConnection.value
    $: if (!$contributorConnection.pending) {
        currentContributorConnection = $contributorConnection.value
    }

    $: timePeriod = data.after

    async function setTimePeriod(event: MouseEvent) {
        const element = event.target as HTMLButtonElement
        timePeriod = element.dataset.value ?? ''
        const newURL = new URL($page.url)
        newURL.search = timePeriod ? `after=${timePeriod}` : ''
        // Don't show stale contributors when switching the time period
        currentContributorConnection = null
        await goto(newURL)
    }
</script>

<svelte:head>
    <title>Contributors - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <div class="root">
        <form method="GET">
            Time period: <input name="after" bind:value={timePeriod} placeholder="All time" />
            <ButtonGroup>
                {#each timePeriodButtons as [label, value]}
                    <Button variant="secondary">
                        <svelte:fragment slot="custom" let:buttonClass>
                            <button
                                class={buttonClass}
                                class:active={timePeriod === value}
                                type="button"
                                data-value={value}
                                on:click={setTimePeriod}>{label}</button
                            >
                        </svelte:fragment>
                    </Button>
                {/each}
            </ButtonGroup>
        </form>
        {#if !currentContributorConnection && $contributorConnection.pending}
            <div class="info">
                <LoadingSpinner />
            </div>
        {:else if currentContributorConnection}
            {@const nodes = currentContributorConnection.nodes}
            <table>
                <tbody>
                    {#each nodes as contributor}
                        {@const commit = contributor.commits.nodes[0]}
                        <tr>
                            <td
                                ><span><Avatar avatar={contributor.person} --avatar-size="1.5rem" /></span>&nbsp;
                                <span>{contributor.person.displayName}</span>
                            </td>
                            <td
                                ><Timestamp date={new Date(commit.author.date)} strict />:
                                <a href={commit.canonicalURL}>{commit.subject}</a></td
                            >
                            <td
                                >{contributor.count}&nbsp;{pluralize(
                                    data.isPerforceDepot ? 'changelist' : 'commit',
                                    contributor.count
                                )}</td
                            >
                        </tr>
                    {:else}
                        <tr>
                            <td colspan="3">
                                <Alert variant="info">No contributors found</Alert>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
            {#if nodes.length > 0}
                <div class="paginator">
                    <Paginator
                        disabled={$contributorConnection.pending}
                        pageInfo={currentContributorConnection.pageInfo}
                    />
                    <small>Total contributors: {currentContributorConnection.totalCount}</small>
                </div>
            {/if}
        {:else if $contributorConnection.error}
            <div class="info">
                <Alert variant="danger">
                    Unable to load contributors: {$contributorConnection.error.message}
                </Alert>
            </div>
        {/if}
    </div>
</section>

<style lang="scss">
    section {
        overflow: auto;
        margin-top: 2rem;
    }

    div.root {
        max-width: 54rem;
        margin-left: auto;
        margin-right: auto;
        margin-bottom: 1rem;
    }

    table {
        border-collapse: collapse;
        width: 100%;
        margin-top: 1rem;
    }

    td {
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color);

        tr:last-child & {
            border-bottom: none;
        }

        span {
            white-space: nowrap;
            text-overflow: ellipsis;
        }
    }

    .paginator {
        display: flex;
        flex-direction: column;
        align-items: center;

        small {
            margin-top: 0.5rem;
            color: var(--text-muted);
        }
    }

    .info {
        margin-top: 1rem;
    }
</style>

<script lang="ts">
    import { goto } from '$app/navigation'
    import { page } from '$app/stores'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { getRelativeTime } from '$lib/relativeTime'
    import UserAvatar from '$lib/UserAvatar.svelte'
    import { currentDate } from '$lib/stores'

    import type { PageData } from './$types'
    import Paginator from '$lib/Paginator.svelte'
    import { Button, ButtonGroup } from '$lib/wildcard'

    export let data: PageData

    const timePeriodButtons = [
        ['Last 7 days', '7 days ago'],
        ['Last 30 days', '30 days ago'],
        ['Last year', '1 year ago'],
        ['All time', ''],
    ]

    $: timePeriod = data.after
    $: contributorsLoader = data.contributors
    $: loading = $contributorsLoader.loading
    let connection: Extract<typeof $contributorsLoader, { loading: false }>['data']
    $: if (!$contributorsLoader.loading && $contributorsLoader.data) {
        connection = $contributorsLoader.data
    }

    async function setTimePeriod(event: MouseEvent) {
        const element = event.target as HTMLButtonElement
        timePeriod = element.dataset.value ?? ''
        const newURL = new URL($page.url)
        newURL.search = timePeriod ? `after=${timePeriod}` : ''
        connection = null
        await goto(newURL)
    }
</script>

<section>
    <div class="container">
        <form method="GET">
            Time period: <input name="after" bind:value={timePeriod} placeholder="All time" />
            <ButtonGroup>
                {#each timePeriodButtons as [label, value]}
                    <Button variant="secondary">
                        <button
                            slot="custom"
                            let:className
                            class={className}
                            class:active={timePeriod === value}
                            type="button"
                            data-value={value}
                            on:click={setTimePeriod}>{label}</button
                        >
                    </Button>
                {/each}
            </ButtonGroup>
        </form>
        {#if !connection && loading}
            <div class="mt-3">
                <LoadingSpinner />
            </div>
        {:else if connection}
            {@const nodes = connection.nodes}
            <table class="mt-3">
                <tbody>
                    {#each nodes as contributor}
                        {@const commit = contributor.commits.nodes[0]}
                        <tr>
                            <td
                                ><span><UserAvatar user={contributor.person} /></span>&nbsp;<span
                                    >{contributor.person.displayName}</span
                                ></td
                            >
                            <td
                                >{getRelativeTime(new Date(commit.author.date), $currentDate)}:
                                <a href={commit.url}>{commit.subject}</a></td
                            >
                            <td>{contributor.count}&nbsp;commits</td>
                        </tr>
                    {/each}
                </tbody>
            </table>
            <div class="d-flex flex-column align-items-center">
                <Paginator disabled={loading} pageInfo={connection.pageInfo} />
                <p class="mt-1 text-muted">
                    <small>Total contributors: {connection.totalCount}</small>
                </p>
            </div>
        {/if}
    </div>
</section>

<style lang="scss">
    section {
        overflow: auto;
        margin-top: 2rem;
    }

    div.container {
        max-width: 54rem;
        margin-left: auto;
        margin-right: auto;
        margin-bottom: 1rem;
    }

    table {
        border-collapse: collapse;
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
</style>

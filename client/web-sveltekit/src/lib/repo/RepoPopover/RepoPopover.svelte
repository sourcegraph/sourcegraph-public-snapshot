<script lang="ts" context="module">
    import type { GraphQLClient } from '$lib/graphql'

    import { RepoPopoverQuery } from './RepoPopover.gql'

    export async function fetchRepoPopoverData(client: GraphQLClient, repoName: string): Promise<RepoPopoverFragment> {
        const response = await client.query(RepoPopoverQuery, { repoName })
        if (!response.data?.repository || response.error) {
            throw new Error(`Failed to fetch repo info: ${response.error}`)
        }
        return response.data.repository
    }
</script>

<script lang="ts">
    import Avatar from '$lib/Avatar.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import Badge from '$lib/wildcard/Badge.svelte'

    import RepoStars from '../RepoStars.svelte'
    import DisplayRepoName from '../shared/DisplayRepoName.svelte'

    import type { RepoPopoverFragment } from './RepoPopover.gql'

    export let data: RepoPopoverFragment
    export let withHeader = false

    $: commit = data.commit
    $: author = commit?.author
</script>

<div class="root">
    {#if withHeader}
        <header>
            <div><DisplayRepoName repoName={data.name} externalLinks={data.externalURLs} /></div>
            <Badge variant="outlineSecondary" small pill>
                {data.isPrivate ? 'Private' : 'Public'}
            </Badge>
        </header>
    {/if}

    {#if data.description || data.topics.length}
        <div class="description-and-tags">
            <div class="description">
                {data.description}
            </div>
            {#if data.topics.length}
                <div class="tags">
                    {#each data.topics as topic}
                        <Badge variant="link" small pill>{topic}</Badge>
                    {/each}
                </div>
            {/if}
        </div>
    {/if}

    {#if commit}
        <div class="last-commit">
            <small>Last Commit</small>

            <div class="commit-info">
                <small class="subject"><a href={commit.canonicalURL}>{commit.subject}</a></small>
                {#if author?.person}
                    <div class="author">
                        <Avatar avatar={author.person} --avatar-size="1.0rem" />
                        <small>{author.person.displayName} · <Timestamp date={author?.date} /></small>
                    </div>
                {/if}
            </div>
        </div>
    {/if}

    <div class="footer">
        <RepoStars repoStars={data.stars} />
    </div>
</div>

<style lang="scss">
    .root {
        width: 480px;

        & > * {
            padding: 0.75rem;

            &:not(:last-child) {
                border-bottom: 1px solid var(--border-color);
            }
        }
    }

    header {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-weight: 500;
    }

    .description-and-tags {
        display: flex;
        flex-flow: column nowrap;
        justify-content: center;
        align-items: flex-start;
        gap: 0.5rem;
        width: 100%;
        color: var(--text-body);

        .tags {
            align-content: space-around;
            align-items: flex-start;
            display: flex;
            flex-flow: row wrap;
            gap: 0.5rem;
            justify-content: flex-start;
        }
    }

    .last-commit {
        display: flex;
        justify-content: space-between;
        gap: 2rem;
        font-size: var(--font-size-small);
        color: var(--text-muted);

        small {
            flex-shrink: 0;
        }

        .commit-info {
            display: flex;
            flex-flow: column nowrap;
            text-align: end;
            align-items: flex-end;
            gap: 0.25rem;
            min-width: 0;

            .subject {
                text-overflow: ellipsis;
                overflow: hidden;
                white-space: nowrap;
                width: 100%;
                a {
                    color: var(--text-body);
                }
            }

            .author {
                display: flex;
                color: var(--text-muted);
                gap: 0.5rem;
                align-items: center;
            }
        }
    }

    .footer {
        display: flex;
        color: var(--text-muted);
        justify-content: flex-end;
        font-size: var(--font-size-tiny);
    }
</style>

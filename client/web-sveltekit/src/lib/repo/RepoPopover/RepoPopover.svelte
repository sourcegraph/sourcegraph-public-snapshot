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
    import Icon from '$lib/Icon.svelte'
    import { displayRepoName } from '$lib/shared'
    import Timestamp from '$lib/Timestamp.svelte'
    import Badge from '$lib/wildcard/Badge.svelte'

    import RepoStars from '../RepoStars.svelte'
    import { getHumanNameForCodeHost, getIconForCodeHost } from '../shared/codehost'

    import type { RepoPopoverFragment } from './RepoPopover.gql'

    export let data: RepoPopoverFragment
    export let withHeader = false

    $: commit = data.commit
    $: author = commit?.author
</script>

<div class="root">
    {#if withHeader}
        <div class="header">
            <div class="left">
                <Icon icon={ILucideGitMerge} aria-hidden --icon-color="var(--primary)" inline />
                <h4>{displayRepoName(data.name)}</h4>
                <Badge variant="outlineSecondary" small pill>
                    {data.isPrivate ? 'Private' : 'Public'}
                </Badge>
            </div>
            <div class="right">
                <Icon
                    icon={getIconForCodeHost(data.externalRepository.serviceType)}
                    --icon-color="var(--text-body)"
                    inline
                />
                <small>{getHumanNameForCodeHost(data.externalRepository.serviceType)}</small>
            </div>
        </div>
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

        & > div {
            padding: 0.5rem 0.75rem;

            &:not(:last-child) {
                border-bottom: 1px solid var(--border-color);
            }
        }
    }

    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .left {
            display: flex;
            justify-content: flex-start;
            align-items: center;
            gap: 0.25rem;

            h4 {
                color: var(--text-title);
                margin: 0;
            }

            small {
                border: 1px solid var(--text-muted);
                color: var(--text-muted);
                padding: 0rem 0.5rem;
                border-radius: 1rem;
            }
        }

        .right {
            display: flex;
            justify-content: flex-end;
            align-items: center;
            gap: 0.25rem;

            small {
                color: var(--text-muted);
            }
        }
    }

    .description-and-tags {
        display: flex;
        flex-flow: column nowrap;
        justify-content: center;
        align-items: flex-start;
        gap: 0.5rem 0.5rem;
        width: 100%;
        color: var(--text-body);

        .tags {
            align-content: space-around;
            align-items: flex-start;
            display: flex;
            flex-flow: row wrap;
            gap: 0.5rem 0.5rem;
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
        font-size: var(--font-size-xs);
    }
</style>

<!--
This Component should be instantiated inside of a Popover component.

For example:

<Popover ...>
    [triggering element ...]
    <div slot="content">
        <RepoPopover ... />
    </div>
</Popover>
-->
<script lang="ts">
    import { mdiSourceMerge } from '@mdi/js'
    import { capitalize } from 'lodash'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import { displayRepoName } from '$lib/shared'
    import Timestamp from '$lib/Timestamp.svelte'
    import Badge from '$lib/wildcard/Badge.svelte'

    import RepoStars from '../RepoStars.svelte'
    import { getIconPathForCodeHost } from '../shared/codehost'

    import type { RepoPopoverFragment } from './RepoPopover.gql'

    export let repo: RepoPopoverFragment
    export let withHeader = false
    export let orgSVGPath: string = mdiSourceMerge

    const CENTER_DOT = '\u00B7' // interpunct

    $: commit = repo.commit
    $: author = commit?.author
    $: avatar = author?.person
    $: codeHostKind = repo.externalServices.nodes[0].kind
</script>

<div class="root">
    {#if withHeader}
        <div class="header">
            <div class="left">
                <Icon svgPath={orgSVGPath} --icon-fill-color="var(--primary)" />
                <h4>{displayRepoName(repo.name)}</h4>
                <Badge variant="outlineSecondary" small pill>
                    {repo.isPrivate ? 'Private' : 'Public'}
                </Badge>
            </div>
            <div class="right">
                <Icon svgPath={getIconPathForCodeHost(codeHostKind)} --icon-fill-color="var(--text-body)" --size={24} />
                <small>{capitalize(codeHostKind)}</small>
            </div>
        </div>

        <div class="divider" />
    {/if}

    <div class="description-and-tags">
        <div class="description">
            {repo.description}
        </div>
        {#if repo.topics.length}
            <div class="tags">
                {#each repo.topics as topic}
                    <Badge variant="link" small pill>{topic}</Badge>
                {/each}
            </div>
        {/if}
    </div>

    <div class="divider" />

    {#if commit}
        <div class="last-commit">
            <small>Last Commit</small>

            <div class="commit-info">
                <div class="commit">
                    <small class="subject">{commit.subject}</small>
                    <small class="commit-number"
                        ><a href={commit.canonicalURL} target="_blank">
                            {commit.abbreviatedOID}
                        </a></small
                    >
                </div>
                {#if author && avatar}
                    <div class="author">
                        <Avatar {avatar} --avatar-size="1.0rem" />
                        <small>{avatar.name}</small>
                        <small>{CENTER_DOT}</small>
                        <small><Timestamp date={author?.date} /></small>
                    </div>
                {/if}
            </div>
        </div>
    {/if}

    <div class="divider" />

    <div class="footer">
        <small>{repo.language}</small>
        <RepoStars repoStars={repo.stars} small={true} />
    </div>
</div>

<style lang="scss">
    .root {
        border: 1px solid var(--dropdown-border-color);
        border-radius: var(--popover-border-radius);
        width: 480px;
    }

    .header {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;
        padding: 0.5rem 0.75rem;
        background-color: var(--subtle-bg);
        border-top-left-radius: var(--border-radius);
        border-top-right-radius: var(--border-radius);

        .left {
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-start;
            align-items: center;
            gap: 0.25rem 0.5rem;

            h4 {
                color: var(--text-body);
                margin-top: 0.5rem;
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
            flex-flow: row nowrap;
            justify-content: flex-end;
            align-items: center;
            gap: 0rem 0.25rem;

            small {
                color: var(--text-muted);
            }
        }
    }

    .divider {
        border-bottom: 1px solid var(--border-color);
        width: 100%;
    }

    .description-and-tags {
        display: flex;
        flex-flow: column nowrap;
        justify-content: center;
        align-items: flex-start;
        gap: 0.5rem 0.5rem;
        padding: 0.75rem;
        width: 100%;

        .description {
            padding: 0rem;
            color: var(--text-body);
        }

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
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: flex-start;
        padding: 0.75rem;

        small {
            color: var(--text-muted);
        }

        .commit-info {
            display: flex;
            flex-flow: column nowrap;
            justify-content: center;
            align-items: flex-end;
            gap: 0.25rem 0rem;

            .commit {
                display: flex;
                flex-flow: row nowrap;
                justify-content: flex-end;
                align-items: center;
                gap: 0.25rem 0.25rem;
                width: 250px;

                .subject {
                    text-overflow: ellipsis;
                    overflow: hidden;
                    white-space: nowrap;
                    color: var(--text-body);
                }

                .commit-number {
                    color: var(--text-muted);
                    font-family: var(--monospace-font-family);
                }
            }

            .author {
                display: flex;
                flex-flow: row nowrap;
                color: var(--text-muted);
                gap: 0.5rem 0.25rem;
            }
        }
    }

    .footer {
        color: var(--text-muted);
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem;
    }
</style>

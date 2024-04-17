<!--
This Component should be instantiated inside of a Popover component.

For example:

<Popover ...>
    [trigger button ...]
    <div slot="content">
        <RepoPopover ... />
    </div>
</Popover>
-->
<script lang="ts">
    import { mdiAlienOutline } from '@mdi/js'
    import { capitalize } from 'lodash'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import Timestamp from '$lib/Timestamp.svelte'

    import RepoStars from '../RepoStars.svelte'
    import { getIconPathForCodeHost } from '../shared/codehost'

    import type { RepoPopoverFragment } from './RepoPopover.gql'

    export let repo: RepoPopoverFragment
    export let withHeader = false

    const CENTER_DOT = '\u00B7' // interpunct

    function truncateCommitNumber(numStr: string | undefined, length: number): string | null {
        if (!numStr) {
            return null
        }
        return numStr.substring(numStr.length - length)
    }

    function formatRepoName(repoName: string): string {
        const slashes = repoName.split('/')
        let repo = slashes.pop()
        let org = slashes.pop()
        return `${org} / ${repo}`
    }

    $: firstFiveTags = repo.tags.nodes.slice(0, 5)
    $: subject = repo.commit?.subject
    $: url = repo.commit?.canonicalURL
    $: author = repo.commit?.author.person.name
    $: commitDate = repo.commit?.author.date
    $: avatar = repo.commit?.author.person
    $: codeHostKind = repo.externalServices.nodes[0].kind
    $: codeHostIcon = getIconPathForCodeHost(codeHostKind)
    $: abbreviatedCommitSHA = truncateCommitNumber(repo.commit?.oid, 6)
</script>

<div class="root">
    {#if withHeader}
        <div class="header">
            <div class="icon-name-access">
                <!-- @TODO: We need to use our customer's logo here. mdiAlienOutline is a place holder-->
                <Icon svgPath={mdiAlienOutline} --color="var(--primary)" />
                <div>
                    <h4 class="repo-name">{formatRepoName(repo.name)}</h4>
                </div>
                <small>{repo.isPrivate ? 'Private' : 'Public'}</small>
            </div>
            <div class="code-host">
                <Icon svgPath={codeHostIcon} --color="var(--text-body)" --size={24} />
                <small>{capitalize(codeHostKind)}</small>
            </div>
        </div>
        <div class="divider" />
    {/if}

    {#if repo.description || firstFiveTags.length > 0}
        <div class="description-and-tags">
            <div class="description">
                {repo.description}
            </div>
            <div class="tags">
                {#each firstFiveTags as tag}
                    <small>{tag.name}</small>
                {/each}
            </div>
        </div>
    {/if}

    <div class="divider" />

    <div class="last-commit">
        <small>Last Commit</small>

        <div class="commit-info">
            <div class="commit">
                <small class="subject">{subject}</small>
                <small class="commit-number">&nbsp;<a href={url} target="_blank">{abbreviatedCommitSHA}</a></small>
            </div>
            <div class="author">
                {#if avatar}
                    <Avatar {avatar} --avatar-size="1.0rem" />
                {/if}
                <small>{author}</small>
                {#if commitDate}
                    <small>{CENTER_DOT}</small>
                    <small><Timestamp date={commitDate} /></small>
                {/if}
            </div>
        </div>
    </div>

    <div class="divider" />

    <div class="footer">
        <small>{repo.language}</small>
        <RepoStars repoStars={repo.stars} small={true} />
    </div>
</div>

<style lang="scss">
    .root {
        width: 480px;
        border: 1px solid var(--border-color);
        border-radius: 0.5rem;
    }

    .header {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;
        padding: 0.5rem 0.75rem;
        background-color: var(--subtle-bg);

        .icon-name-access {
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;
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

        .code-host {
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
        }

        .tags {
            align-content: space-around;
            align-items: flex-start;
            display: flex;
            flex-flow: row wrap;
            gap: 0.5rem 0.5rem;
            justify-content: flex-start;

            small {
                background-color: var(--subtle-bg);
                border-radius: 1rem;
                color: var(--primary);
                font-family: var(--monospace-font-family);
                padding: 0.15rem 0.35rem;
                font-size: 10px;
            }
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
                gap: 0.25rem 0rem;
                width: 250px;

                .subject {
                    text-overflow: ellipsis;
                    overflow: hidden;
                    white-space: nowrap;
                    color: var(--text-body);
                    align-self: center;
                }

                .commit-number {
                    color: var(--text-muted);
                    align-self: center;
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

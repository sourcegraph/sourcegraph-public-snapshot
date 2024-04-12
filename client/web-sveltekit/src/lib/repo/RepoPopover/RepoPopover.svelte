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
    import { capitalize } from 'lodash'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import Timestamp from '$lib/Timestamp.svelte'

    import RepoStars from '../RepoStars.svelte'
    import { getIconPathForCodeHost } from '../shared/codehost'
    import type { RepoPopoverFields } from './RepoPopover.gql'

    export let repo: RepoPopoverFields
    export let withHeader = false

    const CENTER_DOT = '\u00B7' // interpunct

    function truncateCommitNumber(numStr: string, length: number) {
        return numStr.substring(numStr.length - length)
    }

    $: subject = repo.commit?.subject
    $: url = repo.commit?.canonicalURL
    $: commitSHA = repo.commit?.oid
    $: author = repo.commit?.author.person.name
    $: commitDate = repo.commit?.author.date
    $: avatar = repo.commit?.author.person
    $: codeHostKind = repo.externalServices.nodes[0].kind
    $: codeHostIcon = getIconPathForCodeHost(codeHostKind)
</script>

<div class="root">
    {#if withHeader}
        <div class="header">
            <div class="icon-name-access">
                <!-- @TODO: We need to use our customer's logo here, not the code host's -->
                <!--Icon svgPath={mdiGitlab} /-->
                <h4 class="repo-name">{repo.name}</h4>
                <div class="access">
                    <small>{repo.isPrivate ? 'Private' : 'Public'}</small>
                </div>
            </div>
            <div class="code-host">
                <Icon svgPath={codeHostIcon} --color="var(--text-body)" --icon-size="24px" />
                <div><small>{capitalize(codeHostKind)}</small></div>
            </div>
        </div>
        <div class="divider" />
    {/if}

    <div class="description-and-tags">
        <div class="description">{repo.description}</div>
        <div class="tags">
            {#if repo.tags.nodes.length > 0}
                {#each repo.tags.nodes as tag}
                    <div class="tag"><small>{tag.name}</small></div>
                {/each}
            {/if}
        </div>
    </div>

    <div class="divider" />

    <div class="last-commit">
        <div class="heading">
            <small>Last Commit</small>
        </div>

        <div class="commit-info">
            <div class="commit">
                <!--
                A <div> element is needed for subject and commit message
                because the <small> element alone doesn't work with
                text-overflow: ellipsis.
                -->
                <div class="subject">
                    <small>{subject}</small>
                </div>
                {#if commitSHA}
                    <div class="commit-number">
                        <small class="commit-number"
                            ><a href={url} target="_blank">#{truncateCommitNumber(commitSHA, 6)}</a></small
                        >
                    </div>
                {/if}
            </div>
            <div class="author">
                {#if avatar}
                    <Avatar {avatar} --avatar-size="1.0rem" />
                {/if}
                <small>{author}</small>
                <small>{CENTER_DOT}</small>
                {#if commitDate}
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
        border: 1px solid var(--border-color);
        border-radius: var(--popover-border-radius);
        width: 400px;
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

            .repo-name {
                color: var(--text-body);
                margin: 0rem 0.5rem 0rem 0rem;
            }

            .access {
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

            div {
                color: var(--text-muted);
                margin-left: 0.25rem;
            }
        }
    }

    .divider {
        border-bottom: 1px solid var(--border-color);
        width: 100%;
    }

    .description-and-tags {
        padding: 0.75rem;
        width: 100%;

        .description {
            padding: 0rem;
            text-overflow: ellipsis;
            overflow: hidden;
            white-space: nowrap;
        }

        .tags {
            align-content: space-around;
            align-items: flex-start;
            display: flex;
            flex-flow: row wrap;
            gap: 0.5rem 0.5rem;
            justify-content: flex-start;
            margin-top: 0.5rem;

            .tag {
                background-color: var(--subtle-bg);
                border-radius: 1rem;
                color: var(--primary);
                font-family: var(--monospace-font-family);
                padding: 0rem 0.5rem;
            }
        }
    }

    .last-commit {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: flex-start;
        padding: 0.75rem;

        .heading {
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
                }

                .commit-number {
                    color: var(--text-muted);
                    align-self: center;
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

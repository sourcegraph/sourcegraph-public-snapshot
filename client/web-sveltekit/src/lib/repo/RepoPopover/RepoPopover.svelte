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
    import {
        mdiAws,
        mdiBitbucket,
        mdiGit,
        mdiGithub,
        mdiGitlab,
        mdiMicrosoftAzureDevops,
        mdiSourceMerge,
        mdiStarOutline,
    } from '@mdi/js'
    import { formatDistanceToNow } from 'date-fns'
    import { capitalize } from 'lodash'

    import Avatar from '$lib/Avatar.svelte'
    import { ExternalServiceKind } from '$lib/graphql-types'
    import Icon from '$lib/Icon.svelte'

    import { RepoPopoverFields } from './RepoPopover.gql'

    export let repo: RepoPopoverFields
    export let withHeader: Boolean = false

    const CENTER_DOT = '\u00B7' // interpunct

    function formatNumber(num: number): string {
        if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K'
        }
        return num.toString()
    }

    function getCodeHostIcon(kind: ExternalServiceKind): string {
        switch (kind) {
            case ExternalServiceKind.GITHUB:
                return mdiGithub
            case ExternalServiceKind.GITLAB:
                return mdiGitlab
            case ExternalServiceKind.BITBUCKETSERVER:
                return mdiBitbucket
            case ExternalServiceKind.BITBUCKETCLOUD:
                return mdiBitbucket
            case ExternalServiceKind.GITOLITE:
                return mdiGit
            case ExternalServiceKind.AZUREDEVOPS:
                return mdiMicrosoftAzureDevops
            case ExternalServiceKind.AWSCODECOMMIT:
                return mdiAws
            default:
                return mdiSourceMerge
        }
    }

    $: subject = repo.commit?.subject
    $: commitNumber = repo.commit?.abbreviatedOID
    $: author = repo.commit?.author.person.name
    $: commitDate = repo.commit?.author.date
    $: avatar = repo.commit?.author.person
    $: codeHostKind = repo.externalServices.nodes[0].kind
    $: codeHostIcon = getCodeHostIcon(codeHostKind)
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
                <Icon svgPath={codeHostIcon} --color="var(--text-body)" --size={24} />
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
        <div class="title-and-commit">
            <div class="title">
                <small>Last Commit</small>
            </div>
            <div class="commit-and-number">
                <div class="commit">
                    <small>{subject}</small>
                </div>
                {#if commitNumber}
                    <div class="number">
                        <small>{commitNumber}</small>
                    </div>
                {/if}
            </div>
        </div>
        <div class="commit-info">
            <div class="author-and-time">
                {#if avatar}
                    <Avatar {avatar} --avatar-size="1.0rem" />
                {/if}
                <div class="author">
                    <small>{author}</small>
                </div>
                <div class="separator">{CENTER_DOT}</div>

                {#if commitDate}
                    <div class="commit-date">
                        <small>{formatDistanceToNow(commitDate, { addSuffix: false })}</small>
                    </div>
                {/if}
            </div>
        </div>
    </div>
    <div class="divider" />
    <div class="repo-stats">
        <div class="stats">
            <div class="stat"><small>{repo.language}</small></div>
        </div>
        <div class="stat">
            <Icon svgPath={mdiStarOutline} size={16} style="margin-right: 0.15rem;" />
            <small>{formatNumber(repo.stars)}</small>
        </div>
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
        flex-flow: row-nowrap;
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
                // only needed when icon is present
                margin: 0rem 0.5rem 0rem 0rem;
                // border: 1px dotted black;
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
        // border: 1px dotted white;
        padding: 0.75rem;

        .description {
            font-size: 1rem;
            padding: 0rem;
        }

        .tags {
            align-content: space-around;
            align-items: flex-start;
            display: flex;
            flex-flow: row wrap;
            gap: 0.5rem 0rem;
            justify-content: flex-start;
            margin-top: 0.5rem;
        }

        .tag {
            text-align: center;
            background-color: var(--subtle-bg);
            border-radius: 1rem;
            color: var(--primary);
            font-family: var(--monospace-font-family);
            margin-right: 0.5rem;
            padding: 0rem 0.5rem;
        }
    }

    .last-commit {
        display: flex;
        flex-flow: column nowrap;
        justify-content: space-between;
        padding: 0.75rem;

        .title-and-commit {
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;
            align-items: center;

            .title {
                color: var(--text-muted);
            }

            .commit-and-number {
                display: flex;
                flex-flow: row nowrap;
                align-items: center;
                justify-content: flex-end;
                width: 200px;

                .commit {
                    color: var(--text-body);
                    margin-right: 0.25rem;
                    text-overflow: ellipsis;
                    overflow: hidden;
                    white-space: nowrap;
                }

                .number {
                    color: var(--text-muted);
                }
            }
        }

        .author-and-time {
            color: var(--text-muted);
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-end;
            align-items: center;

            .author {
                color: var(--text-muted);
                margin-right: 0.5rem;
                margin-left: 0.5rem;
            }

            .separator {
                margin-right: 0.5rem;
            }
        }
    }

    .repo-stats {
        color: var(--text-muted);
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 0.5rem;
        margin-top: 0.5rem;
        padding: 0rem 0.75rem;

        .stats {
            display: flex;
            flex-flow: row nowrap;
            align-content: center;
            font-size: 1rem;
            padding: 0rem;

            .stat {
                align-self: center;
                margin-right: 1rem;
            }
        }
    }
</style>

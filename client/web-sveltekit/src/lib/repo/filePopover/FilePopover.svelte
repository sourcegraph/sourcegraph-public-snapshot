<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
</script>

<script lang="ts">
    import { mdiFolder, mdiLanguageGo } from '@mdi/js'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import Timestamp from '$lib/Timestamp.svelte'
    import { formatBytes } from '$lib/utils'
    import type { Avatar_Person } from '$testing/graphql-type-mocks'

    import type { FilePopoverFields } from './FilePopover.gql'
    import NodeLine from './NodeLine.svelte'

    faker.seed(1)
    export let repoName: string = 'github.com/sourcegraph/sourcegraph'
    export let f: FilePopoverFields

    interface RepoInfo {
        org: string
        repo: string
    }

    const CENTER_DOT = '\u00B7' // interpunct

    function formatRepoName(repoName: string): RepoInfo {
        const slashes = repoName.split('/')
        let repo = slashes[slashes.length - 1]
        let org = slashes[slashes.length - 2]
        return { org, repo }
    }
    function truncateCommitNumber(numStr: string | undefined, length: number): string | null {
        if (!numStr) {
            return null
        }
        return numStr.substring(numStr.length - length)
    }

    $: ({ repo, org } = formatRepoName(repoName))
    $: filePath = f.path.split('/')
    $: fileOrDirName = filePath.pop()
    $: commit = f.commit
    // TODO: @jasonhawkharris Don't hard code this.
    $: fileInfo = `${f.languages[0]} ${CENTER_DOT} ${f.totalLines} Lines ${CENTER_DOT} ${formatBytes(f.byteSize)}`
    $: dirInfo = `${f.languages[0]} ${CENTER_DOT} 92 Files ${CENTER_DOT} ${formatBytes(f.byteSize)} total size`
    $: commitSHA = truncateCommitNumber(commit.oid, 6)
    $: commitMsg = commit.subject
    $: isDir = f.isDirectory
    $: avatar = commit.author.person
    $: date = commit.author.date
    $: url = commit.canonicalURL

    let team = '@team-code-search'
    let members: Avatar_Person[] = [
        {
            __typename: 'Person',
            displayName: 'Peter Slack',
            name: 'sqs',
            avatarURL: faker.internet.avatar(),
        },
        {
            __typename: 'Person',
            displayName: 'Jason Slack',
            name: 'sqs',
            avatarURL: faker.internet.avatar(),
        },
        {
            __typename: 'Person',
            displayName: 'Michael Slack',
            name: 'sqs',
            avatarURL: faker.internet.avatar(),
        },
        {
            __typename: 'Person',
            displayName: 'Camden Slack',
            name: 'sqs',
            avatarURL: faker.internet.avatar(),
        },
    ]
</script>

<div class="root">
    <div class="desc">
        <div class="repo-and-path">
            <div>{org}</div>
            <div>/</div>
            <div>{repo}</div>
            <div>{CENTER_DOT}</div>
            {#each filePath as part}
                <div>{part}</div>
                {#if filePath.indexOf(part) < filePath.length - 1}
                    <div>/</div>
                {/if}
            {/each}
        </div>

        <div class="lang-and-file">
            <Icon svgPath={isDir ? mdiFolder : mdiLanguageGo} --color="var(--primary)" />
            <div class="file">
                <div>{fileOrDirName}</div>
                <small>{isDir ? dirInfo : fileInfo}</small>
            </div>
        </div>
    </div>

    <div class="last-commit">
        <small class="title">Last Changed @</small>
        <div class="commit">
            <NodeLine />
            <div>
                <a href={url} target="_blank">
                    {commitSHA}
                </a>
                <div class="msg">{commitMsg}</div>
                <div class="author">
                    <Avatar {avatar} --avatar-size="1.0rem" />
                    <small class="name">{avatar.displayName}</small>
                    <small><Timestamp {date} /></small>
                </div>
            </div>
        </div>
    </div>

    <div class="own">
        <div class="own-info">
            <div class="team">Owned by {team}</div>
            <small>{members.length} team members</small>
        </div>
        <div class="members">
            {#each members.slice(0, 5) as member}
                <div class="member">
                    <Avatar avatar={member} --avatar-size="1.0rem" />
                </div>
            {/each}
        </div>
    </div>
</div>

<style lang="scss">
    .root {
        width: fit-content;
        min-width: 380px;
        max-width: 380px;
        background: var(--body-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;

        .desc {
            display: flex;
            flex-flow: column nowrap;
            align-items: center;
            justify-content: center;

            .repo-and-path {
                align-items: center;
                border-bottom: 1px solid var(--border-color);
                display: flex;
                flex-flow: row nowrap;
                gap: 0.25rem;
                justify-content: center;
                padding: 0.5rem 1rem;
                width: 100%;

                div {
                    font-family: var(--monospace-font-family);
                    font-weight: 100;
                    font-size: 0.75rem;
                    color: var(--text-muted);
                }
            }

            .lang-and-file {
                width: 100%;
                display: flex;
                flex-flow: row nowrap;
                align-items: flex-start;
                justify-content: flex-start;
                padding: 0.5rem 1rem;
                gap: 0.25rem 0.75rem;

                .file {
                    display: flex;
                    flex-flow: column nowrap;
                    align-items: flex-start;
                    justify-content: flex-start;
                    font-family: var(--monospace-font-family);
                    gap: 0.25rem;

                    div {
                        color: var(--text-body);
                    }

                    small {
                        color: var(--text-muted);
                    }
                }
            }
        }

        .last-commit {
            display: flex;
            flex-flow: column nowrap;
            align-items: flex-start;
            justify-content: center;
            gap: 0.5rem 0.5rem 0rem;

            .title {
                padding: 0.5rem 1rem;
                color: var(--text-body);
                background-color: var(--secondary-4);
                width: 100%;
                border-bottom: 1px solid var(--border-color);
            }

            .commit {
                padding-left: 1.5rem;
                display: flex;
                flex-flow: row nowrap;
                align-items: center;
                justify-content: flex-start;
                width: 100%;
                height: 90px;
                gap: 0.5rem 1.25rem;

                div {
                    display: flex;
                    flex-flow: column nowrap;
                    align-items: flex-start;
                    justify-content: center;
                    gap: 0.25rem;
                    width: 275px;

                    a {
                        font-family: var(--monospace-font-family);
                        background-color: var(--color-bg-2);
                        padding: 0.15rem 0.25rem;
                        border-radius: 3px;
                        font-size: 0.65rem;
                    }

                    .msg {
                        color: var(--text-body);
                        text-overflow: ellipsis;
                        overflow: hidden;
                        white-space: nowrap;
                    }

                    .author {
                        display: flex;
                        flex-flow: row nowrap;
                        justify-content: flex-start;
                        align-items: center;
                        gap: 0.25rem 0.5rem;
                        color: var(--text-muted);

                        .name {
                            margin-right: 0.5rem;
                        }
                    }
                }
            }
        }
    }

    .own {
        padding: 0.5rem 1rem 0.75rem;
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
        justify-content: space-between;
        border-top: 1px solid var(--border-color);
        color: var(--text-muted);
        .own-info {
            display: flex;
            flex-flow: column nowrap;
            align-items: flex-start;
            justify-content: center;
            gap: 0.25rem;
        }
        .members {
            display: flex;
            flex-flow: row-reverse nowrap;

            .member {
                margin-left: -0.25rem;
            }
        }
    }
</style>

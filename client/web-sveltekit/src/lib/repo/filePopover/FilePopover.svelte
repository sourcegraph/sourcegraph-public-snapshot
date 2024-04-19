<script lang="ts">
    import { mdiLanguageGo } from '@mdi/js'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import type { Avatar_Person } from '$testing/graphql-type-mocks'

    const CENTER_DOT = '\u00B7' // interpunct

    let org = 'sourcegraph'
    let repo = 'sourcegraph'
    let filePath = 'cmd/frontend/auth'.split('/')
    let fileInfo = `Go ${CENTER_DOT} 58 lines ${CENTER_DOT} 1.43 KB`
    let commitSHA = 'def123'
    let commitMsg = 'Adding changes to redis caching to be...'
    let avatar: Avatar_Person = {
        __typename: 'Person',
        displayName: 'Quinn Slack',
        name: 'sqs',
        avatarURL: null,
    }
    let author = 'Quinn Slack'
    let date = '2 days ago'
</script>

<div class="root">
    <div class="desc">
        <div class="repo-and-path">
            <small>{org}</small>
            <small>/</small>
            <small>{repo}</small>
            <small>{CENTER_DOT}</small>
            {#each filePath as part}
                {#if filePath.indexOf(part) < filePath.length - 1}
                    <small>{part}</small>
                    <small>/</small>
                {:else}
                    <small>{part}</small>
                {/if}
            {/each}
        </div>
        <div class="lang-and-file">
            <div>
                <Icon svgPath={mdiLanguageGo} --color="var(--primary)" />
            </div>
            <div class="file">
                <div class="file-name">auth.go</div>
                <div class="file-info">
                    <small>{fileInfo}</small>
                </div>
            </div>
        </div>
    </div>
    <div class="last-commit">
        <div class="title">Last Changed @</div>
        <div class="commit">
            <div class="line" />
            <div class="commit-info">
                <a href="https://github.com/sourcegraph/sourcegraph/commit/{commitSHA}" target="_blank">
                    <small class="sha">{commitSHA}</small>
                </a>
                <div class="msg">{commitMsg}</div>
                <div class="author">
                    <div class="author-info">
                        <Avatar {avatar} --avatar-size="1.0rem" />
                        <small>{author}</small>
                    </div>
                    <small>{date}</small>
                </div>
            </div>
        </div>
    </div>
</div>

<style lang="scss">
    .root {
        width: fit-content;
        background: var(--body-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;

        .desc {
            display: flex;
            flex-flow: column nowrap;
            align-items: center;
            justify-content: center;
            .repo-and-path {
                width: 100%;
                display: flex;
                flex-flow: row nowrap;
                align-items: center;
                justify-content: center;
                gap: 0.25rem;
                padding: 0.5rem 1rem;
                font-family: var(--monospace-font-family);
                font-weight: lighter;
                color: var(--text-muted);
                border-bottom: 1px solid var(--border-color);
            }

            .lang-and-file {
                width: 100%;
                display: flex;
                flex-flow: row nowrap;
                align-items: flex-start;
                justify-content: flex-start;
                padding: 0.5rem 1rem;
                gap: 0.5rem 1rem;
                border-bottom: 1px solid var(--border-color);

                .file {
                    display: flex;
                    flex-flow: column nowrap;
                    align-items: flex-start;
                    justify-content: flex-start;
                    font-family: var(--monospace-font-family);
                    .file-name {
                        color: var(--text-body);
                    }
                    .file-info {
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
                padding-left: 2rem;
                display: flex;
                flex-flow: row nowrap;
                align-items: center;
                justify-content: flex-start;
                width: 100%;
                height: 90px;
                gap: 0.5rem 1.5rem;

                .line {
                    border-left: 1px solid var(--border-color);
                    height: 100%;
                }

                .commit-info {
                    display: flex;
                    flex-flow: column nowrap;
                    align-items: flex-start;
                    justify-content: center;
                    gap: 0.25rem;

                    .sha {
                        font-family: var(--monospace-font-family);
                        background-color: var(--color-bg-2);
                        padding: 0.25rem 0.25rem;
                        border-radius: 3px;
                    }

                    .msg {
                        color: var(--text-body);
                    }

                    .author {
                        display: flex;
                        flex-flow: row nowrap;
                        justify-content: flex-start;
                        align-items: center;
                        gap: 0.25rem 1rem;
                        color: var(--text-muted);

                        .author-info {
                            display: flex;
                            flex-flow: row nowrap;
                            justify-content: flex-start;
                            align-items: center;
                            gap: 0.25rem 0.25rem;
                        }
                    }
                }
            }
        }
    }
</style>

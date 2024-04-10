<script lang="ts">
    import { mdiGithub, mdiStarOutline } from '@mdi/js'
    import { formatDistanceToNow } from 'date-fns'
    import { truncate } from 'lodash'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'
    import Popover from '$lib/Popover.svelte'
    import Button from '$lib/wildcard/Button.svelte'

    import RepoMenu from './RepoMenu.svelte'
    import { RepoPopoverFields } from './RepoPopover.gql'

    export let repo: RepoPopoverFields

    const CENTER_DOT = '\u00B7' // interpunct

    let name = repo.name
    let tags = repo.tags.nodes
    let description = repo.description
    let subject = repo.commit?.subject
    let commitNumber = repo.commit?.abbreviatedOID
    let author = repo.commit?.author.person.name
    let commitDate = repo.commit?.author.date
    let lang = repo.commit?.repository?.language
    let stars = repo.stars
    let url = repo.commit?.canonicalURL
    let avatar = repo.commit?.author.person
    let codeHostIcon = mdiGithub

    /*
    We don't have forks or license information yet.
    We can add them later.
    */
    // let forks
    // let license
    function formatNumber(num: number): string {
        if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K'
        }
        return num.toString()
    }
</script>

<Popover let:registerTrigger let:toggle placement="bottom-start">
    <Button variant="secondary" size="sm" outline>
        <svelte:fragment slot="custom" let:buttonClass>
            <button use:registerTrigger class="{buttonClass} progress-button" on:click={() => toggle()}>
                <RepoMenu {name} {codeHostIcon} />
            </button>
        </svelte:fragment>
    </Button>
    <div slot="content" class="container">
        <div class="description-and-tags">
            <div class="description">{description}</div>
            <div class="tags">
                {#if tags.length > 0}
                    {#each tags as tag}
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
                        <!--
                        Since the popover has a fixed width, we use the lodash
                        truncate() function to truncate the subject to 30 characters.
                        We do this instead of using text-overflow: ellipsis because
                        this adds some unwanted padding to the end of the string,
                        which requires a hacky, less maintanable workaround in the CSS.
                        -->
                        <small>{truncate(subject, { length: 30 })}<small /></small>
                    </div>
                    {#if commitNumber}
                        <div class="number">
                            <a href={url}><small>{commitNumber}</small></a>
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
                            <small>{formatDistanceToNow(commitDate, { addSuffix: true })}</small>
                        </div>
                    {/if}
                </div>
            </div>
        </div>
        <div class="divider" />
        <div class="repo-stats">
            <div class="stats">
                <div class="stat"><small>{lang}</small></div>
                <!-- We don't have forks or license information yet. -->
                <!--div class="stat"><Icon svgPath={mdiSourceMerge} size={14} /><small>{commits}k</small></div-->
            </div>
            <!--div class="license"><small>{license}</small></div-->
            <div class="stat">
                <Icon svgPath={mdiStarOutline} size={16} style="margin-right: 0.15rem;" />
                <small>{formatNumber(stars)}</small>
            </div>
        </div>
    </div>
</Popover>

<style lang="scss">
    .container {
        border-radius: var(--popover-border-radius);
        border: 0rem solid var(--border-color);
        width: 400px;
        padding: 0;
    }

    .divider {
        border-bottom: 1px solid var(--border-color);
        padding: 0.5rem 0.75rem;
        width: 100%;
    }

    .description-and-tags {
        margin-top: 0.5rem;
        padding: 0rem 0.75rem;

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
        margin-top: 0.5rem;
        padding: 0rem 0.75rem;

        .title-and-commit {
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;

            .title {
                color: var(--text-muted);
            }

            .commit-and-number {
                display: flex;
                flex-flow: row nowrap;
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

            /* see note above about forks and license
            .license {
                align-self: center;
            } */
        }
    }
</style>

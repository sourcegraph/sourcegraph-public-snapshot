<script lang="ts">
    import { mdiStarOutline } from '@mdi/js'
    import { formatDistanceToNow } from 'date-fns'

    import Avatar from '$lib/Avatar.svelte'
    import Icon from '$lib/Icon.svelte'

    import { RepoPopoverFields } from './RepoPopover.gql'

    export let repo: RepoPopoverFields

    let tags = repo.tags.nodes
    let description = repo.description
    let subject = repo.commit?.subject
    let commitNumber = repo.commit?.abbreviatedOID
    let author = repo.commit?.author.person.name
    let commitDate = repo.commit?.author.date
    let lang = repo.commit?.repository?.language
    let stars = repo.stars.toString()
    let url = repo.commit?.canonicalURL
    let avatar = repo.commit?.author.person

    /*
    We don't have forks or license information yet.
    We can add them later.
    */
    // let forks
    // let license
    $: console.log(repo)

    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="container">
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
        <div class="title">
            <small>Last Commit</small>
        </div>
        <div class="commit-info">
            <div class="subject-and-commit">
                <div class="subject">
                    <!-- TODO: @jason something strange happens when text-overflow is set to ellipsis.
                it looks like the text cuts off when the ellipses isn't needed and it adds extra padding. -->
                    <small>{subject}<small /></small>
                </div>
                {#if commitNumber}
                    <div class="commit-number">
                        <a href={url}><small>{commitNumber}</small></a>
                    </div>
                {/if}
            </div>
            <div class="author-and-time">
                <Avatar {avatar} --avatar-size="1.0rem" />
                <div class="author">
                    <!-- TODO:@jason add avatar -->
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
        <div class="stat"><Icon svgPath={mdiStarOutline} size={18} /><small>{stars}k</small></div>
    </div>
</div>

<style lang="scss">
    .container {
        border-radius: var(--popover-border-radius);
        border: 1px solid var(--border-color);
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
            align-self: center;
            background-color: var(--subtle-bg);
            border-radius: 1rem;
            color: var(--primary);
            font-family: var(--monospace-font-family);
            justify-self: center;
            margin-right: 0.5rem;
            padding: 0rem 0.25rem;
        }
    }

    .last-commit {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        margin-top: 0.5rem;
        padding: 0rem 0.75rem;

        .title {
            color: var(--text-muted);
        }

        .subject-and-commit {
            align-items: center;
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-end;
            width: 200px;

            .subject {
                color: var(--text-body);
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                margin-right: 0.25rem;
            }

            .commit-number {
                color: var(--text-muted);
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
        margin-bottom: 0.5rem;
        margin-top: 0.5rem;
        padding: 0rem 0.75rem;

        .stats {
            display: flex;
            flex-flow: row nowrap;
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

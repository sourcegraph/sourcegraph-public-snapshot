<script lang="ts">
    import { mdiSourceMerge, mdiStarOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'

    let tags = ['open-source', 'code-search', 'sourcegraph', 'code-intelligence']
    let description = 'Code AI Platform with Code Search & Cody'
    let subject = 'This is a commit message, and it can be very long and contain all sorts of unicode characters.'
    let commitNumber = '#89886'
    let author = 'jasonhawkharris'
    let commitDate = '1 hour ago'
    let lang = 'Go'
    let commits = '1.2'
    let stars = '9.2'
    let license = 'MIT'

    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="container">
    <div class="description-and-tags">
        <div class="description">{description}</div>
        <div class="tags">
            {#each tags as tag}
                <div class="tag"><small>{tag}</small></div>
            {/each}
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
                    <small>{subject}</small>
                </div>
                <div class="commit-number">
                    <small>{commitNumber}</small>
                </div>
            </div>
            <div class="author-and-time">
                <div class="author">
                    <!-- TODO:@jason add avatar -->
                    <small>{author}</small>
                </div>
                <div class="separator">{CENTER_DOT}</div>
                <div class="commit-date"><small>{commitDate}</small></div>
            </div>
        </div>
    </div>
    <div class="divider" />
    <div class="repo-stats">
        <div class="stats">
            <div class="stat"><small>{lang}</small></div>
            <div class="stat"><Icon svgPath={mdiSourceMerge} size={14} /><small>{commits}k</small></div>
            <div class="stat"><Icon svgPath={mdiStarOutline} size={14} /><small>{stars}k</small></div>
        </div>
        <div class="license"><small>{license}</small></div>
    </div>
</div>

<style lang="scss">
    .container {
        border-radius: var(--popover-border-radius);
        border: 1px solid var(--border-color);
        min-width: 400px;
        max-width: 500px;
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

            .subject {
                color: var(--text-body);
                margin-right: 0.5rem;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
                width: 150px;
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

            .author {
                color: var(--text-muted);
                margin-right: 0.5rem;
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

            .license {
                align-self: center;
            }
        }
    }
</style>

<script lang="ts">
    import { mdiSourceMerge, mdiStarOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'

    export let name: string = 'jasonhawkharris'
    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="container">
    <div class="description-and-tags">
        <div class="description">Code AI Platform with Code Search & Cody</div>
        <div class="tags">
            <div class="tag"><small>open-source</small></div>
            <div class="tag"><small>code-search</small></div>
            <div class="tag"><small>sourcegraph</small></div>
            <div class="tag"><small>code-intelligence</small></div>
        </div>
    </div>
    <!-- TODO: add divider-->
    <div class="last-commit">
        <div class="title">
            <small>Last Commit</small>
        </div>
        <div class="commit-info">
            <div class="subject-and-commit">
                <div class="subject">
                    <small>This is a commit message</small>
                </div>
                <div class="commit-number">
                    <small>#89886</small>
                </div>
            </div>
            <div class="author-and-time">
                <div class="author">
                    <!-- TODO:@jason add avatar -->
                    <small>{name}</small>
                </div>
                <div class="separator">{CENTER_DOT}</div>
                <div class="commit-date"><small>1 hour ago</small></div>
            </div>
        </div>
    </div>
    <!-- TODO: add divider-->
    <div class="repo-stats">
        <div class="stats">
            <div class="stat"><small>Go</small></div>
            <div class="stat"><Icon svgPath={mdiSourceMerge} size={14} /><small>1.7k</small></div>
            <div class="stat"><Icon svgPath={mdiStarOutline} size={14} /><small>9.2k</small></div>
        </div>
        <div class="license"><small>MIT</small></div>
    </div>
</div>

<style lang="scss">
    .container {
        border-radius: 6px;
        border: 1px solid var(--border-color);
        min-width: 50px;
        max-width: fit-content; //400px;
        padding: 0.75rem;
        min-height: fit-content;
    }

    .repo-stats {
        padding: 0.5rem 0rem 0rem 0rem;
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        color: var(--text-muted);

        .stats {
            font-size: 1rem;
            padding: 0rem;
            display: flex;
            flex-flow: row nowrap;
            .stat {
                margin-right: 1rem;
                align-self: center;
            }

            .license {
                align-self: center;
            }
        }
    }

    .description-and-tags {
        border-bottom: 1px solid var(--border-color);
        // lists have default margin/padding, so we need to override it
        margin: 0rem;
        padding: 0rem 0rem 0.5rem 0rem;

        .description {
            font-size: 1rem;
            padding: 0rem;
        }

        // @TODO: @jason fix wrapping behavior (spacing between wrapped lines)
        .tags {
            align-items: flex-start;
            display: flex;
            flex-flow: row wrap;
            align-content: space-around;
            justify-content: flex-start;
            margin: 0.5rem 0rem 0.5rem 0rem;
            padding: 0rem;
            line-height: 0.5rem;
        }

        .tag {
            align-self: center;
            background-color: var(--subtle-bg);
            border-radius: 1rem;
            color: var(--primary);
            font-family: var(--monospace-font-family);
            justify-self: center;
            margin-right: 0.5rem;
            padding: 0rem 0.5rem;
        }
    }

    .last-commit {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        padding: 0.5rem 0rem;

        border-bottom: 1px solid var(--border-color);

        .title {
            color: var(--text-muted);
        }

        .subject-and-commit {
            display: flex;
            flex-flow: row nowrap;
            align-items: center;

            .subject {
                color: var(--text-body);
                margin-right: 0.5rem;
            }

            .commit-number {
                color: var(--text-muted);
            }
        }

        .author-and-time {
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-end;
            color: var(--text-muted);

            .author {
                margin-right: 0.5rem;
                color: var(--text-muted);
            }

            .separator {
                margin-right: 0.5rem;
            }
        }
    }
</style>

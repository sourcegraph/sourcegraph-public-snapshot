<script lang="ts">
    export let name: string = 'Jason'
    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="container">
    <div class="description-and-tags">
        <div class="description">Code AI Platform with Code Search & Cody</div>
        <ul class="tags">
            <li class="tag"><small>open-source</small></li>
            <li class="tag"><small>code-search</small></li>
            <li class="tag"><small>sourcegraph</small></li>
            <li class="tag"><small>code-intelligence</small></li>
        </ul>
    </div>
    <!-- TODO: add divider-->
    <div class="last-commit">
        <div><small>Last Commit</small></div>
        <div class="commit-info">
            <div class="subject">
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
        <div class="stats-list">
            <ul class="stats">
                <li class="stat">Go</li>
                <li class="stat">Commits</li>
                <li class="stat">Stars</li>
            </ul>
        </div>
        <div class="license">MIT</div>
    </div>
</div>

<style lang="scss">
    .container {
        border-radius: 10px;
        border: 1px solid var(--border-color);
        min-width: 350px;
        max-width: fit-content;
        padding: 1rem;
    }

    .last-commit {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
    }

    .description-and-tags {
        border-bottom: 1px solid var(--border-color);
        overflow: visible;
        // lists have default margin/padding, so we need to override it
        margin: 0rem;
        padding: 0rem 0rem 0.5rem 0rem;

        .description {
            font-size: 1rem;
            padding: 0rem;
        }

        .tags {
            align-items: flex-start;
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-start;
            list-style-type: none;
            margin: 0.5rem 0rem 0.5rem 0rem;
            padding: 0rem;
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
</style>

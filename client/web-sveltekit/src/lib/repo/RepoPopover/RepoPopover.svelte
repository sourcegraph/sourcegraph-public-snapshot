<script lang="ts">
    export let name: string = 'Jason'
    const CENTER_DOT = '\u00B7' // interpunct
</script>

<div class="container">
    <div class="description-and-tags">
        <div class="description">This is a description of the repo.</div>
        <div class="tag-list">
            <ul class="tags">
                <li class="tag">tag 1</li>
                <li class="tag">tag 2</li>
                <li class="tag">tag 3</li>
                <li class="tag">tag 4</li>
            </ul>
        </div>
    </div>
    <div class="last-commit">
        <div class="subject">
            <div class="subject">this is a commit message</div>
            <div class="commit-number">#89886</div>
        </div>
        <div class="author-and-time">
            <div class="author">{name}</div>
            <div class="separator">{CENTER_DOT}</div>
            <div class="commit-date">1 hour ago</div>
        </div>
    </div>
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
        min-width: 350px;
        max-width: 400px;
        border: 1px solid var(--border-color);
        border-radius: 10px;
        padding: 1rem;
    }
</style>

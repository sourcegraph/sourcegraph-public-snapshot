<article data-testid="search-result">
    <div class="header">
        <div class="icon">
            <slot name="icon" />
        </div>
        <div class="title">
            <slot name="title" />
        </div>
        <div class="info">
            <slot name="info" />
        </div>
    </div>
    {#if $$slots.default || $$slots.body}
        <slot name="body">
            <div class="body">
                <slot />
            </div>
        </slot>
    {/if}
</article>

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        padding: 0.5rem 0.5rem 0.5rem 0;
        position: sticky;
        top: 0;
        background-color: var(--body-bg);
    }

    .icon {
        flex-shrink: 0;
    }

    .title {
        flex: 1 1 auto;
        overflow: hidden;
        display: flex;
        flex-wrap: wrap;

        // .title-inner
        overflow-wrap: anywhere;

        // .muted-repo-file-link
        color: var(--text-muted);

        :global(a) {
            color: var(--text-muted);

            &:hover {
                color: var(--text-muted);
            }
        }
    }

    .info {
        margin-left: auto;
        display: flex;
        flex-wrap: wrap;
    }

    .body {
        border-radius: var(--border-radius);
        border: 1px solid var(--border-color);
        background-color: var(--code-bg);
    }
</style>

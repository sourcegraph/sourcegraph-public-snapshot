<article data-show-copy-target data-testid="search-result">
    <div class="header">
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
    article {
        :global([data-copy-button]) {
            opacity: 0;
            transition: opacity 0.2s;
        }
        &:is(:hover, :focus-within) :global([data-copy-button]) {
            opacity: 1;
        }
    }

    .header {
        display: flex;
        align-items: center;
        padding: 0.5rem 0.75rem;
        position: sticky;
        top: 0;
        background-color: var(--search-result-header-bg);
        border-bottom: 1px solid var(--border-color);
    }

    .title {
        flex: 1 1 auto;
        overflow: hidden;
        display: flex;
        flex-wrap: wrap;
        align-items: center;

        // .title-inner
        overflow-wrap: anywhere;
    }

    .info {
        margin-left: auto;
        display: flex;
        flex-wrap: wrap;
        color: var(--text-muted);
        flex-shrink: 0;
        gap: 0.5rem;
        align-items: center;
    }

    .body:not(:empty) {
        background-color: var(--code-bg);
        border-bottom: 1px solid var(--border-color);
    }
</style>

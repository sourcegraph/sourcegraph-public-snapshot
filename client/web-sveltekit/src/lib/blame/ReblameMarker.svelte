<script lang="ts">
    import { page } from '$app/stores'
    import { SourcegraphURL } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import { getURLToFileCommit, type BlameHunk } from '$lib/web'

    export let hunk: BlameHunk

    $: previous = hunk.commit.previous
    $: href = previous
        ? SourcegraphURL.from(getURLToFileCommit($page.url.href, previous.filename, previous.rev))
              .setLineRange({
                  line: hunk.startLine,
              })
              .toString()
        : null
</script>

{#if href}
    <a {href} title="Reblame prior to {hunk.rev.slice(0, 7)}">
        <Icon icon={ILucideFileStack} aria-hidden inline />
    </a>
{/if}

<style lang="scss">
    a {
        display: flex;
        align-items: center;
        padding: 0 0.125rem;
        height: 100%;

        &:hover {
            --icon-color: currentColor;
            color: var(--text-body);
        }
    }
</style>

<script lang="ts">
    import { getButtonClassName } from '$lib/wildcard/Button'
    import Toggle from '$lib/wildcard/Toggle.svelte'

    export let handleOptOut: (() => void) | undefined
</script>

<div class="root">
    <div class="section header">
        <h4 class="m0">New, faster UX (Beta)</h4>
        {#if handleOptOut}
            <Toggle on={true} on:click={() => handleOptOut()} />
        {/if}
    </div>
    <div class="section">
        <p>Have feedback for us on the beta? We’d love to hear from you.</p>
        <a
            class={getButtonClassName({ variant: 'secondary' })}
            href="https://community.sourcegraph.com/c/code-search/9"
            target="_blank"
            rel="noreferrer noopener"
        >
            Leave feedback
        </a>
        <p class="small">It only takes two minutes and helps a ton!</p>
    </div>
</div>

<style lang="scss">
    .root {
        width: 360px;

        .section {
            padding: 1rem;
            & + .section {
                border-top: 1px solid var(--border-color);
            }
        }

        .m0 {
            margin: 0;
        }

        .header {
            display: flex;
            justify-content: space-between;
        }

        a {
            width: 100%;
        }

        .small {
            color: var(--text-muted);
            font-size: var(--font-size-xs);
            text-align: center;
            margin-top: 0.25rem;
            margin-bottom: 0;
        }
    }
</style>

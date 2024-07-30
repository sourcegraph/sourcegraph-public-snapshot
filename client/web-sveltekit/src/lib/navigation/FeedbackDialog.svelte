<script lang="ts">
    import { getButtonClassName } from '$lib/wildcard/Button'
    import Toggle from '$lib/wildcard/Toggle.svelte'

    export let handleOptOut: (() => void) | undefined
</script>

<div class="root">
    <div class="section header">
        <h4>New, faster UX (Beta)</h4>
        {#if handleOptOut}
            <Toggle on={true} on:click={() => handleOptOut()} />
        {/if}
    </div>
    <div class="section">
        <p>Got feedback for us on the beta? We’d love to hear from you.</p>
        <a class={getButtonClassName({ variant: 'secondary' })} href="https://community.sourcegraph.com/">
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

        .header {
            display: flex;
            justify-content: space-between;
            h4 {
                margin: 0;
            }
        }

        a {
            width: 100%;
        }

        .small {
            color: var(--text-muted);
            font-size: var(--font-size-extra-small);
            text-align: center;
            margin-top: 0.25rem;
            margin-bottom: 0;
        }
    }
</style>

<script lang="ts">
    export let links: { name: string; href: string; handleClick: () => void }[]
</script>

{#if links.length > 0}
    <footer>
        {#each links as link}
            <a href={link.href} on:click={link.handleClick} rel="noopener noreferrer" target="_blank">
                {link.name}
            </a>
        {/each}
    </footer>
{/if}

<style lang="scss">
    footer {
        display: flex;
        flex-direction: row;

        a {
            color: var(--text-muted);
            padding: 0 1rem;
            &:not(:last-child) {
                border-right: 1px solid var(--border-color);
            }
        }

        // In a small viewport, align links in a column and remove the separator
        @media (--xs-breakpoint-down) {
            flex-direction: column;
            gap: 0.5rem;
            align-items: center;

            a:not(:last-child) {
                border: none;
            }
        }
    }
</style>

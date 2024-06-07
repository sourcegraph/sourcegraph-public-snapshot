<script lang="ts">
    export let links: { name: string; href: string }[]
</script>

{#if links.length > 0}
    <footer>
        {#each links as link}
            <a href={link.href}>{link.name}</a>
        {/each}
    </footer>
{/if}

<style lang="scss">
    footer {
        display: flex;
        flex-direction: column;

        a {
            color: var(--text-muted);
        }
    }
</style>

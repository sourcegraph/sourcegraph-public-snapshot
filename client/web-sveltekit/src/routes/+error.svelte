<script lang="ts">
    import { page } from '$app/stores'
</script>

<div>
    {#if $page.error?.message === 'enterprise feature'}
        <h1>Enterprise feature</h1>
        <p>This is an enterprise-only feature. See ... for more information.</p>
    {:else}
        <h1>{$page.status}</h1>
        <p>
            {$page.error?.message}
        </p>
    {/if}
</div>

<style lang="scss">
    div {
        margin: 1rem;
    }
</style>

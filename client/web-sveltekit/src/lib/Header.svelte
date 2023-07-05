<script lang="ts">

    import UserAvatar from '$lib/UserAvatar.svelte'

    import { user } from './stores'

    $: authenticatedUser = user
</script>

<header>
    <div class="header">
        <slot />
    </div>
    <!--
    <div class="user">
        {#if $authenticatedUser}
            <UserAvatar user={$authenticatedUser} />
            <!--
                Needs data-sveltekit-reload to force hitting the server and
                proxying the request to the Sourcegraph instance.
            - ->
            <a href="/-/sign-out" data-sveltekit-reload>Sign out</a>
        {:else}
            <a href="/sign-in" data-sveltekit-reload>Sign in</a>
        {/if}
    </div>
    -->
</header>


<style lang="scss">
    .header {
        display: flex;
        align-self: center;
        flex: 1;
    }

    header {
        display: flex;
        align-items: center;
        border-bottom: 1px solid var(--border-color-2);
        height: var(--navbar-height);
        min-height: 40px;
        padding: 0 1rem;
        background-color: var(--color-bg-1);
    }

    ul {
        position: relative;
        padding: 0;
        margin: 0;
        list-style: none;
        background-size: contain;
    }
</style>

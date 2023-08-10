<script lang="ts">
    interface User {
        avatarURL?: string | null
        displayName?: string | null
        username?: string | null
    }

    export let user: User

    function getInitials(name: string): string {
        const names = name.split(' ')
        const initials = names.map(name => name.charAt(0).toLowerCase())
        if (initials.length > 1) {
            return `${initials[0]}${initials[initials.length - 1]}`
        }
        return initials[0]
    }

    $: name = user.displayName || user.username || ''
</script>

{#if user.avatarURL}
    <img src={user.avatarURL} role="presentation" aria-hidden="true" alt="Avatar of {name}" />
{:else}
    <div>
        <span>{getInitials(name)}</span>
    </div>
{/if}

<style lang="scss">
    img,
    div {
        isolation: isolate;
        display: inline-flex;
        border-radius: 50%;
        text-transform: capitalize;
        color: var(--color-bg-1);
        align-items: center;
        justify-content: center;
        min-width: 1.5rem;
        min-height: 1.5rem;
        position: relative;
        background: linear-gradient(to bottom, var(--logo-purple), var(--logo-orange));
        width: var(--avatar-size, var(--icon-inline-size));
        height: var(--avatar-size, var(--icon-inline-size));
    }

    div::after {
        content: '';
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        left: 0;
        border-radius: 50%;
        background: linear-gradient(to right, var(--logo-purple), var(--logo-blue));
        mask-image: linear-gradient(to bottom, #000000, transparent);
    }

    span {
        z-index: 1;
        color: var(--white);
    }
</style>

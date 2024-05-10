<script lang="ts">
    import type { Avatar_User, Avatar_Team, Avatar_Person } from './Avatar.gql'

    type Avatar = Avatar_User | Avatar_Team | Avatar_Person

    export let avatar: Avatar

    function getInitials(name: string): string {
        const names = name.split(' ')
        const initials = names.map(name => name.charAt(0).toLowerCase())
        if (initials.length > 1) {
            return `${initials[0]}${initials[initials.length - 1].toUpperCase()}`
        }
        return initials[0]
    }

    function getName(avatar: Avatar): string {
        switch (avatar.__typename) {
            case 'User':
                return avatar.displayName || avatar.username || ''
            case 'Person':
                return avatar.displayName || avatar.name || ''
            case 'Team':
                return avatar.displayName || ''
        }
    }

    $: name = getName(avatar)
    $: avatarURL = avatar.avatarURL
</script>

{#if avatarURL}
    <img src={avatarURL} role="presentation" aria-hidden="true" alt="Avatar of {name}" data-avatar />
{:else}
    <div data-avatar>
        <span>{getInitials(name)}</span>
    </div>
{/if}

<style lang="scss">
    span {
        z-index: 1;
        color: var(--text-muted);
        font-size: calc(var(--size) * 0.5);
        font-weight: 500;
    }

    img,
    div {
        --min-size: 1.25rem;
        --size: var(--avatar-size, var(--icon-inline-size, var(--min-size)));

        min-width: var(--min-size);
        min-height: var(--min-size);
        width: var(--size);
        height: var(--size);
        isolation: isolate;
        display: inline-flex;
        border-radius: 50%;
        text-transform: capitalize;
        color: var(--color-bg-1);
        align-items: center;
        justify-content: center;
        position: relative;
        background: var(--secondary);
    }

    div::after {
        content: '';
        position: absolute;
        top: 0;
        right: 0;
        bottom: 0;
        left: 0;
        border-radius: 50%;
    }
</style>

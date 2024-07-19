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
    <div data-avatar title={name}>
        <span>{getInitials(name)}</span>
    </div>
{/if}

<style lang="scss">
    span {
        color: var(--text-muted);
        font-size: calc(var(--size) * 0.5);
        font-weight: 500;
    }

    img,
    div {
        --min-size: 1.25rem;
        --size: var(--avatar-size, var(--icon-inline-size, var(--min-size)));

        flex: none;

        min-width: var(--min-size);
        min-height: var(--min-size);
        width: var(--size);
        height: var(--size);
        border-radius: 50%;

        display: inline-flex;
        align-items: center;
        justify-content: center;

        text-transform: capitalize;
        color: var(--color-bg-1);
        background: var(--secondary);
        user-select: none;
    }
</style>

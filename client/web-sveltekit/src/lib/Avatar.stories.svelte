<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import type { Avatar_Person } from './Avatar.gql'
    import Avatar from './Avatar.svelte'

    export const meta = {
        component: Avatar,
    }
</script>

<script lang="ts">
    faker.seed(1)

    const avatar: Avatar_Person = {
        // Having __typename is relevant here because getName() uses it
        // to determine the initials of a person to display in the Avatar.
        __typename: 'Person',
        avatarURL: faker.internet.avatar(),
        displayName: 'Quinn Slack',
        name: faker.internet.userName(),
    }
</script>

<Story name="Default">
    <div class="root">
        <Avatar {avatar} />
        <Avatar avatar={{ ...avatar, avatarURL: null }} />
        <Avatar {avatar} --avatar-size="1.5rem" />
        <Avatar avatar={{ ...avatar, avatarURL: null }} --avatar-size="1.5rem" />
        <Avatar {avatar} --avatar-size="2.5rem" />
        <Avatar avatar={{ ...avatar, avatarURL: null }} --avatar-size="2.5rem" />
        <Avatar {avatar} --avatar-size="3.5rem" />
        <Avatar avatar={{ ...avatar, avatarURL: null }} --avatar-size="3.5rem" />
    </div>
</Story>

<style lang="scss">
    .root {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-evenly;
        align-items: center;
        gap: 0.5rem 1.25rem;
        width: fit-content;
    }
</style>

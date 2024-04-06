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
        displayName: `${faker.person.firstName()} ${faker.person.lastName()}`,
        name: faker.internet.userName(),
    }
</script>

<Story name="Default">
    <h2>With <code>avatarURL</code>and default size</h2>
    <Avatar {avatar} />
    <h2>With <code>avatarURL</code>and custom size</h2>
    <Avatar {avatar} --avatar-size="3rem" />
    <h2>Without <code>avatarURL</code>and default size</h2>
    <Avatar avatar={{ ...avatar, avatarURL: null }} />
    <h2>Without <code>avatarURL</code>and custom size</h2>
    <Avatar avatar={{ ...avatar, avatarURL: null }} --avatar-size="6rem" />
    <h2>Without <code>avatarURL</code>and huge size</h2>
    <Avatar avatar={{ ...avatar, avatarURL: null }} --avatar-size="15rem" />
</Story>

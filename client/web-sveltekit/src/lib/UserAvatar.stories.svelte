<script lang="ts" context="module">
    import { Story } from '@storybook/addon-svelte-csf'
    import UserAvatar from './UserAvatar.svelte'
    import { faker } from '@faker-js/faker'

    export const meta = {
        component: UserAvatar,
    }
</script>

<script lang="ts">
    faker.seed(1)
    const avatarURL = faker.internet.avatar()
    const username = faker.internet.userName()
    const displayName = `${faker.person.firstName()} ${faker.person.lastName()}`
</script>

<Story name="Default">
    <h2>With <code>avatarURL</code></h2>
    <UserAvatar user={{ avatarURL }} />
    <h2>With <code>username</code> "{username}"</h2>
    <UserAvatar user={{ username }} />
    <h2>With <code>displayName</code> "{displayName}"</h2>
    <UserAvatar user={{ displayName }} />
</Story>

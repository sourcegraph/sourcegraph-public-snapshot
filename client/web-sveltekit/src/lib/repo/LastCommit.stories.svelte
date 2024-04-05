<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import type { LastCommitFragment } from './LastCommit.gql'
    import LastCommit from './LastCommit.svelte'

    export const meta = {
        component: LastCommit,
    }
</script>

<script lang="ts">
    faker.seed(1)

    const withAvatar: LastCommitFragment = {
        id: faker.datatype.number.toString(),
        abbreviatedOID: faker.datatype.number.toString(),
        subject: faker.lorem.sentence(),
        canonicalURL: faker.internet.url(),
        author: {
            date: new Date().toISOString(),
            person: {
                __typename: 'Person',
                displayName: `${faker.person.firstName()} ${faker.person.lastName()}`,
                name: faker.internet.userName(),
                avatarURL: faker.internet.avatar(),
            },
        },
    }

    const withoutAvatar: LastCommitFragment = {
        id: faker.datatype.number.toString(),
        abbreviatedOID: faker.datatype.number.toString(),
        subject: faker.lorem.sentence(),
        canonicalURL: faker.internet.url(),
        author: {
            date: new Date().toISOString(),
            person: {
                __typename: 'Person',
                displayName: `${faker.person.firstName()} ${faker.person.lastName()}`,
                name: faker.internet.userName(),
                avatarURL: null,
            },
        },
    }
</script>

<Story name="Default">
    <h2>With <code>avatarURL</code></h2>
    <LastCommit lastCommit={withAvatar} />
    <br />
    <h2>Without <code>avatarURL</code></h2>
    <LastCommit lastCommit={withoutAvatar} />
</Story>

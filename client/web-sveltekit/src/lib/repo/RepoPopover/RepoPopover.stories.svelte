<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import { type RepoPopoverFields, ExternalServiceKind } from '$testing/graphql-type-mocks'

    import RepoPopover from './RepoPopover.svelte'

    let withHeader = true

    export const meta = {
        component: RepoPopover,
    }
</script>

<script lang="ts">
    faker.seed(1)
    let repo: RepoPopoverFields = {
        name: `${faker.lorem.word()} / ${faker.lorem.word()}`,
        description: faker.lorem.sentence(),
        stars: faker.datatype.number(),
        tags: {
            nodes: [
                { name: faker.lorem.word() },
                { name: faker.lorem.word() },
                { name: faker.lorem.word() },
                { name: faker.lorem.word() },
            ],
        },
        isPrivate: false,
        language: 'Go',
        externalServices: {
            totalCount: 1,
            nodes: [
                {
                    kind: ExternalServiceKind.GITHUB,
                },
            ],
        },
        commit: {
            id: faker.datatype.number.toString(),
            subject: faker.lorem.sentence(),
            canonicalURL: faker.internet.url(),
            oid: '7b4d3ad230d9078a70219f2befa1be1fe00377a0',
            author: {
                date: new Date().toISOString(),
                person: {
                    __typename: 'Person',
                    displayName: `${faker.person.firstName()} ${faker.person.lastName()}`,
                    avatarURL: faker.internet.avatar(),
                    name: faker.internet.userName(),
                },
            },
        },
    }
</script>

<Story name="Default">
    <h2>With header</h2>
    <RepoPopover {repo} withHeader />
    <br />
    <br />
    <h2>Without header</h2>
    <RepoPopover {repo} />
</Story>

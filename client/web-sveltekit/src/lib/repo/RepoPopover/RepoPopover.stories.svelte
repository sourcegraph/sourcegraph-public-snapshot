<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import { type RepoPopoverFields, type CodeHostFields, ExternalServiceKind } from '$testing/graphql-type-mocks'

    import RepoPopover from './RepoPopover.svelte'

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
        commit: {
            id: faker.datatype.number.toString(),
            subject: faker.lorem.sentence(),
            abbreviatedOID: '#87873',
            canonicalURL: faker.internet.url(),
            author: {
                date: new Date().toISOString(),
                person: {
                    displayName: `${faker.person.firstName()} ${faker.person.lastName()}`,
                    avatarURL: faker.internet.avatar(),
                    name: faker.internet.userName(),
                },
            },
            repository: {
                language: 'Go',
            },
        },
    }
    let codeHost: CodeHostFields = {
        kind: ExternalServiceKind.GITHUB,
    }
</script>

<Story name="Default">
    <h2>With header</h2>
    <RepoPopover {repo} {codeHost} withHeader={true} />
    <br />
    <br />
    <h2>Without header</h2>
    <RepoPopover {repo} {codeHost} />
</Story>

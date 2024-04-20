<script lang="ts" context="module">
    import { faker } from '@faker-js/faker'
    import { Story } from '@storybook/addon-svelte-csf'

    import type { FilePopoverFields } from '$testing/graphql-type-mocks'

    import FilePopover from './FilePopover.svelte'

    export const meta = {
        component: FilePopover,
    }
</script>

<script lang="ts">
    faker.seed(1)
    let withFile: FilePopoverFields = {
        path: `${faker.lorem.word()}/${faker.lorem.word()}/${faker.lorem.word()}/${faker.lorem.word()}.go`,
        isDirectory: false,
        languages: ['Go'],
        totalLines: 4539,
        byteSize: 239945873,
        commit: {
            oid: 'dk83mf98392n39v963y2n03n',
            subject: faker.lorem.sentence(),
            canonicalURL: faker.internet.url(),
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
    let withDir: FilePopoverFields = {
        path: `${faker.lorem.word()}/${faker.lorem.word()}/${faker.lorem.word()}/${faker.lorem.word()}.go`,
        isDirectory: true,
        languages: ['Go'],
        totalLines: 4539,
        byteSize: 239945873,
        commit: {
            oid: 'dk83mf98392n39v963y2n03n',
            subject: faker.lorem.sentence(),
            canonicalURL: faker.internet.url(),
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
    <h4>File Popover</h4>
    <FilePopover repoName="sourcegraph/sourcegraph" f={withFile} />
    <br />
    <br />
    <h4>Dir Popover</h4>
    <FilePopover repoName="sourcegraph/sourcegraph" f={withDir} />
</Story>

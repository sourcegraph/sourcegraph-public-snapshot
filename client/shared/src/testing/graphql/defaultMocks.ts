import { fakerEN as faker } from '@faker-js/faker'

export const defaultMocks: Record<string, () => unknown> = {
    // GraphQLs Int type is 32 bit so we need to set the max value.
    Int: () => faker.number.int(2 ** 31),
    String: () => faker.string.alpha(10),
    ID: () => faker.string.uuid(),
    // Ensure some simple business logic

    // In Ranges the start line is always smaller than the end line.
    Range: () => {
        const line = faker.number.int(1000)
        return {
            start: {
                line,
                character: faker.number.int(100),
            },
            end: {
                line: line + faker.number.int(50),
                character: faker.number.int(100),
            },
        }
    },

    Person: () => ({
        avatarURL: faker.image.avatar(),
        email: faker.helpers.arrayElement([faker.internet.email(), faker.internet.email(), '']),
    }),

    UserEmail: () => ({
        email: faker.internet.email(),
    }),
}

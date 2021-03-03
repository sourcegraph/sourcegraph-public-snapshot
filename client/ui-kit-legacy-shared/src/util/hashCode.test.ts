import { hashCode } from './hashCode'

describe('hashCode', () => {
    const testCaseUUIDs = {
        'd3e406d5-a359-4b88-b0d3-db025a957811': 3,
        'd3f07658-2629-45ad-9159-86b738f0b6bf': 18,
        'd905e503-24da-4983-b537-14652745e625': 9,
        'dc53d94d-d7cc-4038-bdd0-a7331c034484': 5,
        'e2574dd3-80f3-4a98-b8fc-73616edc90f5': 9,
        'e59b8eaf-d04d-43cd-bf0b-a6af3154e44e': 8,
        'e9f3ee64-b64d-473d-9c17-d89071cbd223': 6,
        'f2969c9b-b6bf-456b-8969-d38d837c791a': 9,
        'f579f9db-02f5-434a-8af5-2dac5fe44780': 0,
        'f5eab38b-cb20-45b2-b3f2-55531239ca35': 10,
        'f693c26b-7689-40b3-8852-2b38d9295f6c': 12,
        'f7c1f6b3-eecd-44fb-b38d-6942e730ebc4': 8,
        'fcffc897-0c62-433c-ae7c-b5f5f2933f6c': 4,
        'fe84ef0e-c8ec-42ee-ba03-45a017fda5e6': 9,
    }

    test('UUIDs', () => {
        for (const [UUID, hash] of Object.entries(testCaseUUIDs)) {
            expect(hashCode(UUID, 20)).toBe(hash)
        }
    })

    test('is deterministic', () => {
        const extensionID = 'some-website.co/publisher/name'

        const hash = hashCode(extensionID, 20)

        expect(hashCode(extensionID, 20)).toBe(hash)
    })
})

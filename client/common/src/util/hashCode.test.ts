import { describe, expect, test } from '@jest/globals'

import { hashCode } from './hashCode'

describe('hashCode', () => {
    const testCaseUUIDs = {
        'd3e406d5-a359-4b88-b0d3-db025a957811': 'EvvIjqiG1H71c2DNHbO4m2E/hvN2cYpGVRH+hcLW4uM=',
        'd3f07658-2629-45ad-9159-86b738f0b6bf': 'k+mqeb0pK+eT3TbvrR2IEsyhty5NJHKNeNr9AdmlwMc=',
        'd905e503-24da-4983-b537-14652745e625': 'RFM1icFQlBb6frvITgzRrfw0YPUVVPoh4mLmijZ0fns=',
        'dc53d94d-d7cc-4038-bdd0-a7331c034484': '25QwZCtq3vShER0zcdE6jqRyucL9rR2mM5i1Pul2W/s=',
        'e2574dd3-80f3-4a98-b8fc-73616edc90f5': 'nqFZYF04+neLmW+45H7CEulJJwB+AiIKjRGWNqWGzdM=',
        'e59b8eaf-d04d-43cd-bf0b-a6af3154e44e': '0CmRVkLq7SAnsBhihAXVVR3St1F2rVKPgC1gf4u93Do=',
        'e9f3ee64-b64d-473d-9c17-d89071cbd223': 'uAVZyzZ+kAoPzPVsSIs+tR/GMxGu3Nf1ywjZfxlNpUg=',
        'f2969c9b-b6bf-456b-8969-d38d837c791a': 'nTqGsPDpO7D1qbGNMUh6azTANaBoAirx7NuTVMBjsSk=',
        'f579f9db-02f5-434a-8af5-2dac5fe44780': 'LJy7cVFm7jt0OTJ16+S4JQWiXSOf9UakkqK8KEC7clU=',
        'f5eab38b-cb20-45b2-b3f2-55531239ca35': 'bVyUs1nwnjYq1qi52npgyIa2yylwNFkC+Gvy3kzGX4Y=',
        'f693c26b-7689-40b3-8852-2b38d9295f6c': 'gOqAKP/ZWlrftAOotLyqErcD+zIGMnHFvNaq0Egho9E=',
        'f7c1f6b3-eecd-44fb-b38d-6942e730ebc4': 'ejmHOHb57QAbVLBvaXDa7JpJ+MGDZ4f171XWESQb8AM=',
        'fcffc897-0c62-433c-ae7c-b5f5f2933f6c': 'Y70Mopkjt14ip99tekAbMTU+HT0ATOQ7ZeUfQYltsv0=',
        'fe84ef0e-c8ec-42ee-ba03-45a017fda5e6': 'QXasxWcoUgFBh1yLFa6WMSZJ7f7t9XIyg5Wc0rqNC1Y=',
    }

    test('UUIDs', async () => {
        for (const [UUID, hash] of Object.entries(testCaseUUIDs)) {
            expect(await hashCode(UUID)).toBe(hash)
        }
    })

    test('is deterministic', async () => {
        const extensionID = 'some-website.co/publisher/name'

        const hash = await hashCode(extensionID)

        expect(await hashCode(extensionID)).toBe(hash)
    })
})

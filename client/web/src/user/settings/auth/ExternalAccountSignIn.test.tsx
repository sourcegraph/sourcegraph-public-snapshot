import { Attribute, getUsernameOrEmail, SamlExternalData } from './ExternalAccountsSignIn'

function toAttribute(value: string): Attribute {
    return {
        Values: [
            {
                Value: value,
            },
        ],
    }
}

function samlDataObject(keysValues: object): SamlExternalData {
    // Add some other fields to make sure we are getting the right ones
    Object.assign(keysValues, { 'any field': 'false' })
    Object.assign(keysValues, { 'random field': 'Mon Oct 10 2022 13:07:34 GMT+0000 (Coordinated Universal Time)' })
    Object.assign(keysValues, { 'one more random field': 'banana' })

    const testData: unknown = {
        Values: Object.fromEntries(Object.entries(keysValues).map(([key, value]) => [key, toAttribute(value)])),
    }

    return testData as SamlExternalData
}

describe('getUsernameOrEmail', () => {
    test('saml account data has only email', () => {
        const testCases = [
            samlDataObject({
                'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress': 'mary@boole.com',
            }),
            samlDataObject({
                emailaddress: 'mary@boole.com',
            }),
            samlDataObject({
                'http://schemas.xmlsoap.org/claims/EmailAddress': 'mary@boole.com',
            }),
        ]

        for (const testCase of testCases) {
            expect(getUsernameOrEmail(testCase)).toEqual('mary@boole.com')
        }
    })

    test('saml account data has email and username - username should be used', () => {
        const testCases = [
            samlDataObject({
                emailaddress: 'emmy@noether.com',
                username: 'emmynoether',
            }),
            samlDataObject({
                'http://schemas.xmlsoap.org/claims/EmailAddress': 'emmy@noether.com',
                username: 'emmynoether',
            }),
        ]

        for (const testCase of testCases) {
            expect(getUsernameOrEmail(testCase)).toEqual('emmynoether')
        }
    })

    test('saml account data has only username', () => {
        const testCases = [
            samlDataObject({
                'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name': 'adalovelace',
            }),
            samlDataObject({
                nickname: 'adalovelace',
            }),
            samlDataObject({
                login: 'adalovelace',
            }),
            samlDataObject({
                username: 'adalovelace',
            }),
        ]

        for (const testCase of testCases) {
            expect(getUsernameOrEmail(testCase)).toEqual('adalovelace')
        }
    })
})

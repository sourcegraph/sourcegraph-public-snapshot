import { gitHubAppConfig } from './externalServices'

describe('gitHubAppConfig', () => {
    test('get config', () => {
        const config = gitHubAppConfig('https://test.com', '1234', '5678', 'testUser')
        expect(config.defaultConfig).toEqual(`{
  "url": "https://test.com",
  "gitHubAppDetails": {
    "installationID": 5678,
    "appID": 1234,
    "baseURL": "https://test.com"
  },
  "orgs": ["testUser"],
  "authorization": {}
}`)
    })
})

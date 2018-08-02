import assert from 'assert'
import { ConfigurationSubject, ConfiguredSubject } from '../settings'
import { filterItems } from './ExtensionConfigureButton'

const FIXTURE_CONFIGURATION_SUBJECT: ConfigurationSubject = {
    id: '',
    __typename: 'User',
    username: 'n',
    displayName: 'n',
    viewerCanAdminister: true,
    settingsURL: 'a',
}

describe('filterItems', () =>
    it('filters to added only', () =>
        assert.deepStrictEqual(
            filterItems<ConfigurationSubject>(
                'a',
                [
                    { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '1' }, settings: { extensions: { a: true } } },
                    { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '2' }, settings: { extensions: { a: false } } },
                    { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '3' }, settings: { extensions: { b: true } } },
                    { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '4' }, settings: null },
                    { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '4' }, settings: {} },
                ],
                { added: true }
            ),
            [
                { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '1' }, settings: { extensions: { a: true } } },
                { subject: { ...FIXTURE_CONFIGURATION_SUBJECT, id: '2' }, settings: { extensions: { a: false } } },
            ] as ConfiguredSubject<ConfigurationSubject>[]
        )))

import { replaceVersion } from './replaceDependencyVersion'

const GROUP = 'a.b'
const NAME = 'c'
const OLD_VERSION = '1.2'
const NEW_VERSION = '3.4'

describe('replaceDependencyVersion', () => {
    const TESTS: {
        body: string
        want: string
    }[] = [
        { body: `compile "a.b:c:1.2"`, want: `compile "a.b:c:3.4"` },
        { body: `compile('a.b:c:1.2') { force = true}`, want: `compile('a.b:c:3.4') { force = true}` },
    ]
    for (const { body, want } of TESTS) {
        it(body, () =>
            expect(
                replaceVersion(body, {
                    group: GROUP,
                    name: NAME,
                    oldVersion: OLD_VERSION,
                    newVersion: NEW_VERSION,
                })
            ).toBe(want)
        )
    }
})

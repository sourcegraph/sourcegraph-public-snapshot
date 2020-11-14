import { eraseTopLevelParameter } from './transformer'

describe('erases parameter', () => {
    test('erases case:yes', () => {
        expect(eraseTopLevelParameter('repo:sg/sg case:yes SauceGraph', 'case')).toMatchInlineSnapshot(
            `
            Object {
              "type": "success",
              "value": "repo:sg/sg  SauceGraph",
            }
        `
        )
    })

    test('does not erase nested case', () => {
        expect(eraseTopLevelParameter('repo:sg/sg (case:yes SauceGraph)', 'case')).toMatchInlineSnapshot(
            `
            Object {
              "reason": "can only transform parameter at the toplevel, this one occurs in a grouped subexprsesion",
              "type": "error",
            }
        `
        )
    })

    test('does not erase double patterntype', () => {
        expect(
            eraseTopLevelParameter('patterntype:literal SauceGraph or patterntype:regexp wafflecat', 'patterntype')
        ).toMatchInlineSnapshot(
            `
            Object {
              "reason": "can only transform query with one patterntype",
              "type": "error",
            }
        `
        )
    })
})

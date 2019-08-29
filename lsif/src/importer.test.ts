import { reachableMonikers, normalizeHover, lookupRanges } from "./importer"
import { Id } from 'lsif-protocol'

describe('database', () => {
  describe('helpers', () => {
    test('lookupRanges', async () => {
      const ranges = new Map<Id, number>()
      ranges.set(1, 0)
      ranges.set(2, 2)
      ranges.set(3, 1)

      const orderedRanges = [
        { startLine: 1, startCharacter: 1, endLine: 1, endCharacter: 2, monikers: [] },
        { startLine: 2, startCharacter: 1, endLine: 2, endCharacter: 2, monikers: [] },
        { startLine: 3, startCharacter: 1, endLine: 3, endCharacter: 2, monikers: [] },
      ]

      const document = {
        id: "",
        path: "",
        contains: [],
        definitions: [],
        references: [],
        ranges,
        orderedRanges,
        resultSets: new Map(),
        definitionResults: new Map(),
        referenceResults: new Map(),
        hovers: new Map(),
        monikers: new Map(),
        packageInformation: new Map(),
      }

      expect(lookupRanges(document, [1, 2, 3, 4])).toEqual([
        orderedRanges[0],
        orderedRanges[2],
        orderedRanges[1],
      ])
    })

    test('normalizeHover', async () => {
      expect(normalizeHover({ contents: "foo" })).toEqual("foo")
      expect(normalizeHover({ contents: { language: "typescript", value: "bar" } })).toEqual("```typescript\nbar\n```")
      expect(normalizeHover({ contents: { kind: 'markdown', value: "baz" } })).toEqual("baz")
      expect(normalizeHover({
        contents: [
          "foo",
          { language: "typescript", value: "bar" },
        ]
      })).toEqual('foo\n\n---\n\n```typescript\nbar\n```')
    })

    test('reachableMonikers', async () => {
      const monikerSets = new Map<Id, Set<Id>>()
      monikerSets.set(1, new Set<Id>([2]))
      monikerSets.set(2, new Set<Id>([1, 4]))
      monikerSets.set(3, new Set<Id>([4]))
      monikerSets.set(4, new Set<Id>([2, 3]))
      monikerSets.set(5, new Set<Id>([6]))
      monikerSets.set(6, new Set<Id>([5]))

      expect(reachableMonikers(monikerSets, 1)).toEqual(new Set<Id>([
        1, 2, 3, 4
      ]))
    })
  })
})

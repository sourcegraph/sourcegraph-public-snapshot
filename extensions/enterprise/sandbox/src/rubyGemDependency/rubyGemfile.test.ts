jest.mock('sourcegraph', () => ({
    Range: require('@sourcegraph/extension-api-classes').Range,
    Position: require('@sourcegraph/extension-api-classes').Position,
}))

import { RubyGemfileDependency, parseRubyGemfileLock } from './rubyGemfile'
import { readFileSync } from 'fs'
import { resolve } from 'path'

interface SimplifiedRubyGemfileDependency
    extends Pick<RubyGemfileDependency, Exclude<keyof RubyGemfileDependency, 'range' | 'directAncestors'>> {
    range?: string
    directAncestors: string[]
}

// From github.com/mbleigh/acts-as-taggable-on.
const SAMPLE_GEMFILE_LOCK = readFileSync(resolve(__dirname, 'testdata', 'Gemfile.lock.sample'), 'utf-8')
const WANT_GEMFILE_LOCK_SAMPLE: SimplifiedRubyGemfileDependency[] = JSON.parse(
    readFileSync(resolve(__dirname, 'testdata', 'parseRubyGemfileLock.want.json'), 'utf-8')
)

const simplify = (deps: RubyGemfileDependency[]): SimplifiedRubyGemfileDependency[] => {
    return deps.map(dep => {
        const o = {
            ...dep,
            range:
                'range' in dep && dep.range
                    ? `${dep.range.start.line}:${dep.range.start.character}-${dep.range.end.line}:${dep.range.end.character}`
                    : undefined,
        }
        if (o.range === undefined) {
            delete o.range
        }
        return o
    })
}

describe('parseRubyGemfileLock', () => {
    test('simple', () =>
        expect(
            simplify(
                parseRubyGemfileLock(`
PATH
  remote: .
  specs:
    abc (6.0.1)

GEM
  remote: https://rubygems.org/
  specs:
    foo (1.2.3)
      bar (~> 2.3.4)
      baz (= 3.4.5, >= 4.5.6)
    bar (2.3.4)
      baz (= 3.4.5)
    baz (3.4.5)

PLATFORMS
  ruby
  
DEPENDENCIES
	foo

BUNDLED WITH
  1.17.3
`)
            )
        ).toEqual([
            {
                directAncestors: ['foo'],
                name: 'bar',
                range: '11:4-11:15',
                version: '2.3.4',
            },
            {
                directAncestors: ['bar', 'foo'],
                name: 'baz',
                range: '13:4-13:15',
                version: '3.4.5',
            },
            {
                directAncestors: [],
                name: 'foo',
                range: '8:4-8:15',
                version: '1.2.3',
            },
        ] as SimplifiedRubyGemfileDependency[]))
    test('complex', () => expect(simplify(parseRubyGemfileLock(SAMPLE_GEMFILE_LOCK))).toEqual(WANT_GEMFILE_LOCK_SAMPLE))
})

// require('fs').writeFileSync(
//     resolve(__dirname, 'testdata', 'parseRubyGemfileLock.want.json'),
//     JSON.stringify(simplify(parseRubyGemfileLock(SAMPLE_GEMFILE_LOCK)), null, 2)
// )

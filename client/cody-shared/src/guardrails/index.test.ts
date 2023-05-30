import { summariseAttribution } from '.'

const genRepos = (count: number) => {
    const repos = []
    for (let i = 0; i < count; i++) {
        repos.push({ name: `repo${i}` })
    }
    return repos
}

describe('summariseAttribution', () => {
    it('handles error', () => {
        expect(summariseAttribution(new Error('test'))).toEqual('guardrails attribution search failed: test')
    })
    it('handles no matches', () => {
        expect(summariseAttribution({ limitHit: false, repositories: [] })).toEqual('no matching repositories found')
    })
    it('handles one match', () => {
        expect(summariseAttribution({ limitHit: false, repositories: genRepos(1) })).toEqual(
            'found 1 matching repository repo0'
        )
    })
    it('handles five matches', () => {
        expect(summariseAttribution({ limitHit: false, repositories: genRepos(5) })).toEqual(
            'found 5 matching repositories repo0, repo1, repo2, repo3, repo4'
        )
    })
    it('handles many matches', () => {
        expect(summariseAttribution({ limitHit: false, repositories: genRepos(10) })).toEqual(
            'found 10 matching repositories repo0, repo1, repo2, repo3, repo4, ...'
        )
    })
    it('handles many matches limithit', () => {
        expect(summariseAttribution({ limitHit: true, repositories: genRepos(10) })).toEqual(
            'found 10+ matching repositories repo0, repo1, repo2, repo3, repo4, ...'
        )
    })
})

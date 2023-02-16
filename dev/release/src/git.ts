import execa from 'execa'
import { SemVer } from 'semver'
import * as semver from 'semver'

import { localSourcegraphRepo } from './github'

export function getTags(workdir: string, prefix?: string): string[] {
    execa.sync('git', ['fetch', '--tags'], { cwd: workdir })
    return execa
        .sync('git', ['--no-pager', 'tag', '-l', `${prefix}`, '--sort=v:refname'], { cwd: workdir })
        .stdout.split('\n')
}

export function getCandidateTags(workdir: string, version: string): string[] {
    return getTags(workdir, `v${version}-rc*`)
}

export function getReleaseTags(workdir: string): string[] {
    const raw = getTags(workdir, 'v[0-9]*.[0-9]*.[0-9]*')
    // since tags are globbed they can overmatch when we just want pure release tags
    return raw.filter(tag => tag.match('[0-9]+\\.[0-9]+\\.[0-9]+$'))
}

// Returns the version tagged in the repository previous to a provided input version. If no input version it will
// simply return the highest version found in the repository.
export function getPreviousVersion(version?: SemVer): SemVer {
    const lowest = new SemVer('0.0.1')
    const tags = getReleaseTags(localSourcegraphRepo)
    if (tags.length === 0) {
        return lowest
    }
    if (!version) {
        return new SemVer(tags[tags.length - 1])
    }

    for (
        let reallyLongVariableNameBecauseESLintRulesAreSilly = tags.length - 1;
        reallyLongVariableNameBecauseESLintRulesAreSilly >= 0;
        reallyLongVariableNameBecauseESLintRulesAreSilly--
    ) {
        const tag = tags[reallyLongVariableNameBecauseESLintRulesAreSilly]
        const temp = semver.parse(tag)
        if (temp && temp.compare(version) === -1) {
            return temp
        }
    }
    return lowest
}

import { spawnSync } from 'child_process'

describe('summarizer', () => {
    it('summarizes text', () => {
        const proc = spawnSync(
            'git',
            ['log', '--after="1 day ago"', '--format="%an%x09%ae%x09%H"', 'origin/main', '--', ':/'],
            { cwd: '/home/beyang/src/github.com/sourcegraph/sourcegraph' }
        )
        if (proc.status !== 0) {
            throw new Error(`error running git: ${proc.stderr.toString()}`)
        }

        // NEXT
        console.log('# output', proc.stdout.toString().trim())
    })
})

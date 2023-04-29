import { spawn } from 'child_process'

import { CommitNode } from './summarizer'

describe('summarizer', () => {
    it('summarizes text', async () => {
        // git log --after="1 week ago" -p --pretty=format:'<commit>%H</commit><authorName>%an</authorName><authorEmail>$ae</authorEmail><message>%B</message>'
        const child = spawn(
            'git',
            [
                'log',
                '--after=1 day ago',
                '-p',
                '--pretty=format:<commit>%H</commit><authorName>%an</authorName><authorEmail>%ae</authorEmail><message>%B</message>',
            ],
            { cwd: '/home/beyang/src/github.com/sourcegraph/sourcegraph' }
        )
        const out = await new Promise<string>((resolve, reject) => {
            let stdout = ''
            child.stdout.on('data', data => {
                stdout += data
            })

            child.on('close', code => {
                if (code === 0) {
                    resolve(stdout)
                } else {
                    reject(new Error(`git failed with code ${code}`))
                }
            })
        })
        // console.log('# out:', out)
        const nodes = CommitNode.fromText(out)
        console.log('# nodes', nodes)

        // if (proc.status !== 0) {
        //     console.error(proc.stdout.toString())
        //     throw new Error(`error running git (status ${proc.status}): ${proc.stderr.toString()}`)
        // }
        // console.log('# output', proc.stdout.toString().trim())

        // const proc = spawnSync(
        //     'git',
        //     ['log', '--after="1 day ago"', '--format="%an%x09%ae%x09%H"', 'origin/main', '--', ':/'],
        //     { cwd: '/home/beyang/src/github.com/sourcegraph/sourcegraph' }
        // )
        // if (proc.status !== 0) {
        //     throw new Error(`error running git: ${proc.stderr.toString()}`)
        // }

        // console.log('# output', proc.stdout.toString().trim())

        // const nodes = CommitNode.fromText(proc.stdout.toString().trim())
        // console.log('# nodes', nodes)

        // const summarizer = new Summarizer(nodes)
    })
})

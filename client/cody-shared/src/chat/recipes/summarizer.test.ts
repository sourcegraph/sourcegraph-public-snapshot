import { spawn } from 'child_process'

import { SourcegraphBrowserCompletionsClient } from '../../sourcegraph-api/completions/browserClient'
import { ChatClient } from '../chat'

import { CommitNode, parseFileDiffs } from './summarizer'

function firstLastLineSummarize(text: string): Promise<string> {
    const lines = text.split('\n').filter(line => line.length > 0)
    if (lines.length === 0) {
        return Promise.resolve('')
    }
    return Promise.resolve(lines[0] + '\n' + lines[lines.length - 1])
}

describe('summarizer', () => {
    it('summarizes text', async () => {
        jest.setTimeout(10000000)
        // git log --after="1 week ago" -p --pretty=format:'<commit>%H</commit><authorName>%an</authorName><authorEmail>$ae</authorEmail><message>%B</message>'
        const child = spawn(
            'git',
            [
                'log',
                '--after=2 days ago',
                '-p',
                '--pretty=format:<commitMetadata><commit>%H</commit><authorName>%an</authorName><authorEmail>%ae</authorEmail><message>%B</message></commitMetadata>',
                'origin/main',
                '--',
                './',
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
        const nodes = CommitNode.fromText(out)

        const completionsClient = new SourcegraphBrowserCompletionsClient({
            serverEndpoint: 'https://sourcegraph.com',
            accessToken: 'sgp_4973ca15878c144c3166805d5dc6471577ec5a2f',
            debug: false,
            customHeaders: {},
        })
        const summaries = await Promise.all(nodes.map(node => node.getSummary(firstLastLineSummarize)))
        console.log('# summaries', summaries)

        // const llmClient = new ChatClient(completionsClient)
        // const llmSummarize = (text: string): Promise<string> => {
        //     return new Promise((resolve, reject) => {
        //         let summary = ''
        //         llmClient.chat(
        //             [
        //                 {
        //                     speaker: 'human',
        //                     text: `Summarize this:\n${text}`,
        //                 },
        //             ],
        //             {
        //                 onChange: (text: string) => {
        //                     summary += text
        //                 },
        //                 onComplete: () => resolve(summary),
        //                 onError: (message: string, statusCode?: number) =>
        //                     reject(`error ${statusCode && `(${statusCode})`}: ${message}`),
        //             }
        //         )
        //     })
        // }

        // const summaries = await Promise.all(nodes.map(node => node.getSummary(llmSummarize)))
        // console.log('# summaries', summaries.join('\n\n'))
    })

    it('parses file diffs', () => {
        const rawDiff = `diff --git a/client/cody-shared/src/chat/recipes/summarizer.test.ts b/client/cody-shared/src/chat/recipes/summarizer.test.ts
index a5aa8fd8d7..6d0fcc83fe 100644
--- a/client/cody-shared/src/chat/recipes/summarizer.test.ts
+++ b/client/cody-shared/src/chat/recipes/summarizer.test.ts
@@ -1,17 +1,58 @@
-import { spawnSync } from 'child_process'
+import { spawn } from 'child_process'
+
+import { CommitNode } from './summarizer'

 describe('summarizer', () => {
-    it('summarizes text', () => {
-        const proc = spawnSync(
+    it('summarizes text', async () => {
`
        const fileDiffs = parseFileDiffs(rawDiff)
        expect(JSON.stringify(fileDiffs)).toEqual(
            JSON.stringify([
                {
                    oldFilename: 'client/cody-shared/src/chat/recipes/summarizer.test.ts',
                    newFilename: 'client/cody-shared/src/chat/recipes/summarizer.test.ts',
                    diff: `@@ -1,17 +1,58 @@
-import { spawnSync } from 'child_process'
+import { spawn } from 'child_process'
+
+import { CommitNode } from './summarizer'

 describe('summarizer', () => {
-    it('summarizes text', () => {
-        const proc = spawnSync(
+    it('summarizes text', async () => {
`,
                },
            ])
        )
    })
})

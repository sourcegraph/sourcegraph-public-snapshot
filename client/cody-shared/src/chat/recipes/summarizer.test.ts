import { spawn } from 'child_process'

import { CommitNode, Summarizer, parseFileDiffs } from './summarizer'

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

        const summarizer = new Summarizer(nodes)

        // NEXT: write a LLM summarize function, pass it to Summarizer constructor
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

interface SummaryNode {
    getSourceIdentifier(): string
    getText(): Promise<string>
    getReferences(): { name: string; node: SummaryNode }[]
}

interface Commit {
    authorName: string
    authorEmail: string
    hash: string
    rawDiff: string
    diff: FileDiff[]
}

interface FileDiff {
    oldFilename: string
    newFilename: string
    diff: string
}

export class CommitNode implements SummaryNode {
    constructor(private commit: Commit) {}

    public static fromText(text: string): CommitNode[] {
        const commits: Commit[] = []
        const lines = text.split('\n')
        let commit: Commit | undefined
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i]
            console.log(`<line>${line}</line>`)
            if (line.startsWith('<commit>')) {
                if (commit) {
                    commit.diff = parseFileDiffs(commit.rawDiff)
                    commits.push(commit)
                }
                commit = {
                    authorEmail: textBetween(line, '<authorEmail>', '</authorEmail>'),
                    authorName: textBetween(line, 'authorName>', '</authorName>'),
                    hash: textBetween(line, '<commit>', '</commit>'),
                    rawDiff: '',
                    diff: [],
                }
                continue
            }
            if (!commit) {
                continue
            }
            commit.rawDiff += line + '\n'
        }
        if (commit) {
            commit.diff = parseFileDiffs(commit.rawDiff)
            commits.push(commit)
        }
        return commits.map(commit => new CommitNode(commit))
    }

    // public static fromText(text: string): CommitNode[] {
    //     const commitNodes: CommitNode[] = []
    //     const lines = text.split('\n')
    //     for (const line of lines) {
    //         const fields = line.split('\t')
    //         if (fields.length !== 3) {
    //             console.error(`ignoring line because did not find 3 fields: ${line}`)
    //             continue
    //         }
    //         const [authorName, authorEmail, commitHash] = fields
    //         commitNodes.push(new CommitNode(authorName, authorEmail, commitHash))
    //     }
    //     return commitNodes
    // }

    getSourceIdentifier(): string {
        return this.commit.hash
    }
    getText(): Promise<string> {
        throw new Error('Method not implemented.')
    }
    getReferences(): { name: string; node: SummaryNode }[] {
        throw new Error('Method not implemented.')
    }

    private fetchCommitInfo() {
        // commit message
        // file diffs
    }
}

function textBetween(text: string, startTag: string, endTag: string): string {
    const start = text.indexOf(startTag) + startTag.length
    const end = text.indexOf(endTag, start)
    if (start === -1 || end === -1) {
        return ''
    }
    return text.slice(start, end)
}

// Parse a raw diff like this:
// diff --git a/client/cody-shared/src/chat/recipes/summarizer.test.ts b/client/cody-shared/src/chat/recipes/summarizer.test.ts
// index a5aa8fd8d7..6d0fcc83fe 100644
// --- a/client/cody-shared/src/chat/recipes/summarizer.test.ts
// +++ b/client/cody-shared/src/chat/recipes/summarizer.test.ts
// @@ -1,17 +1,58 @@
// -import { spawnSync } from 'child_process'
// +import { spawn } from 'child_process'
// +
// +import { CommitNode } from './summarizer'

//  describe('summarizer', () => {
// -    it('summarizes text', () => {
// -        const proc = spawnSync(
// +    it('summarizes text', async () => {
export function parseFileDiffs(rawDiff: string): FileDiff[] {
    const lines = rawDiff.split('\n')
    const fileDiffs: FileDiff[] = []
    let fileDiff: FileDiff | undefined
    for (const line of lines) {
        if (line.startsWith('diff --git a/')) {
            if (fileDiff) {
                fileDiffs.push(fileDiff)
            }
            fileDiff = {
                oldFilename: '',
                newFilename: '',
                diff: '',
            }
            continue
        }
        if (!fileDiff) {
            continue
        }

        if (line.startsWith('index ')) {
            // skip
        } else if (line.startsWith('--- a/')) {
            fileDiff.oldFilename = line.slice('--- a/'.length)
        } else if (line.startsWith('+++ b/')) {
            fileDiff.newFilename = line.slice('+++ b/'.length)
        } else {
            if (fileDiff.diff.length > 0) {
                fileDiff.diff += '\n'
            }
            fileDiff.diff += line
        }
    }
    if (fileDiff) {
        fileDiffs.push(fileDiff)
    }
    return fileDiffs
}

export class Summarizer {
    constructor(private nodes: SummaryNode[], private summarize: (text: string) => Promise<string>) {}
    getSummary(): string {
        return ''
    }

    private async getSummaryHelper(node: SummaryNode): Promise<string> {
        const refs = node.getReferences()
        if (refs.length === 0) {
            return node.getText()
        }
        const summaries = await Promise.all(
            refs.map(async ref => ({
                summary: await this.getSummaryHelper(ref.node),
                name: ref.name,
            }))
        )
        const summariesText = summaries.map(({ summary, name }) => `${name}: ${summary}`)
        const nodeText = await node.getText()
        const fullText = `${nodeText}\n\n${summariesText.join('\n\n')}`
        return this.summarize(fullText)
    }
}

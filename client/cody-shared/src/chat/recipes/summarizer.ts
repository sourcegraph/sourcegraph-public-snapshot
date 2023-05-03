interface SummaryNode {
    getSummary(summarize: (text: string) => Promise<string>): Promise<string>
}

interface Commit {
    authorName: string
    authorEmail: string
    hash: string
    message: string
    rawDiff: string
    diff: FileDiff[]
}

interface FileDiff {
    oldFilename: string
    newFilename: string
    diff: string
}

export class FileDiffNode implements SummaryNode {
    constructor(private diff: FileDiff) {}
    async getSummary(summarize: (text: string) => Promise<string>): Promise<string> {
        if (this.diff.oldFilename === this.diff.newFilename) {
            return summarize(`${this.diff.oldFilename}:\n${this.diff.diff.trim()}`)
        }
        const filenameText = `${this.diff.oldFilename} became ${this.diff.newFilename}`
        if (this.diff.diff.trim().length === 0) {
            return filenameText
        }
        return `${filenameText}:\n${this.diff.diff.trim()}`
    }
}

export class CommitNode implements SummaryNode {
    constructor(private commit: Commit) {}

    public static fromText(text: string): CommitNode[] {
        const commits: Commit[] = []

        while (text.length > 0) {
            const [commitMetadataText, diffIndex] = textBetween(text, '<commitMetadata>', '</commitMetadata>')
            text = text.substring(diffIndex)
            let diffEndIndex = text.indexOf('<commitMetadata>')
            if (diffEndIndex === -1) {
                diffEndIndex = text.length
            }
            const diffText = text.substring(0, diffEndIndex)
            text = text.substring(diffEndIndex)
            commits.push({
                authorEmail: textBetween(commitMetadataText, '<authorEmail>', '</authorEmail>')[0],
                authorName: textBetween(commitMetadataText, '<authorName>', '</authorName>')[0],
                hash: textBetween(commitMetadataText, '<commit>', '</commit>')[0],
                message: textBetween(commitMetadataText, '<message>', '</message>')[0],
                rawDiff: diffText,
                diff: parseFileDiffs(diffText),
            })
        }
        return commits.map(commit => new CommitNode(commit))
    }

    async getSummary(summarize: (text: string) => Promise<string>): Promise<string> {
        const fileSummaries = await Promise.all(
            this.commit.diff.map(fileDiff => new FileDiffNode(fileDiff).getSummary(summarize))
        )

        return summarize(`${this.commit.message}\n\n${fileSummaries.join('\n\n')}`)
    }
}

function textBetween(text: string, startTag: string, endTag: string): [string, number] {
    const start = text.indexOf(startTag) + startTag.length
    const end = text.indexOf(endTag, start)
    if (start === -1 || end === -1) {
        return ['', -1]
    }
    return [text.slice(start, end), end]
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

// export class Summarizer {
//     constructor(private nodes: SummaryNode[], private summarize: (text: string) => Promise<string>) {}

//     async getSummary(): Promise<string> {
//         const nodeSummaries = await Promise.all(this.nodes.map(node => this.getSummaryHelper(node)))
//         return nodeSummaries.join('\n')
//     }

//     private async getSummaryHelper(node: SummaryNode): Promise<string> {
//         const refs = node.getReferences()
//         // Base case
//         if (refs.length === 0) {
//             return this.summarize(node.getText())
//         }

//         // Recursive case
//         const summaries = await Promise.all(
//             refs.map(async ref => ({
//                 summary: await this.getSummaryHelper(ref.node),
//                 name: ref.name,
//             }))
//         )
//         const summariesText = summaries.map(({ summary, name }) => `${name}: ${summary}`)
//         const nodeText = node.getText()
//         const fullText = `${nodeText}\n\n${summariesText.join('\n\n')}`
//         return this.summarize(fullText)
//     }
// }

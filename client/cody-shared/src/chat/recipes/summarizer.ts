interface SummaryNode {
    getText(): string
    getReferences(): SummaryNode[]
}

// class CommitNode implements SummaryNode {
//     public static fromText(text: string): {
//         return new CommitNode()
//     }
//     constructor() {

//     }
//     getReferences(): SummaryNode[] {
//         throw new Error('Method not implemented.')
//     }
// }

// class Summarizer {
//     constructor(private nodes: SummaryNode[]) {}
// }

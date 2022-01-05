import Parser from 'web-tree-sitter'

import { Indexer, LsifDocument, Input, LsifDocumentBuilder, LsifHighlight } from './lsif'

export class GolangIndexer extends Indexer {
    constructor() {
        super('go')
    }
    public async index(input: Input): Promise<LsifDocument> {
        const tree = await this.parseInput(input)
        console.log({ tree })
        const builder = new LsifDocumentBuilder()
        this.walk(builder, tree.rootNode)
        return {
            occurrences: builder.occurrences,
        }
    }

    private walk(builder: LsifDocumentBuilder, node: Parser.SyntaxNode): void {
        switch (node.type) {
            case 'interpreted_string_literal':
                builder.pushHighlight(node, LsifHighlight.STRING_LITERAL)
                break
            case 'float_literal':
            case 'int_literal':
                builder.pushHighlight(node, LsifHighlight.NUMERIC_LITERAL)
                break
            default:
                break
        }
        for (let i = 0; i < node.childCount; i++) {
            const child = node.child(i)
            if (child !== null) {
                this.walk(builder, child)
            }
        }
    }
    // private next(cursor: Parser.TreeCursor): Parser.SyntaxNode | undefined {
    //     const hasNext = cursor.gotoFirstChild() || cursor.gotoNextSibling() || cursor.gotoParent()
    //     return hasNext ? cursor.currentNode() : undefined
    // }
}

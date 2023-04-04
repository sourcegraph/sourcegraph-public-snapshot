import { Node, Operator, Parameter, Pattern, OperatorKind, Sequence } from './parser'
import { CharacterRange, PatternKind } from './token'

export class Visitor {
    /**
     * Set up this visitor with visit callback functions.
     *
     * @param visitors the visitor callback functions.
     */
    constructor(private readonly _visitors: Visitors) {}

    /**
     * Top-level visit function.
     *
     * @param nodes Top-level nodes of the tree.
     */
    public visit(nodes: Node[]): void {
        for (const node of nodes) {
            switch (node.type) {
                case 'operator':
                    this.visitOperator(node)
                    break
                case 'sequence':
                    this.visitSequence(node)
                    break
                case 'parameter':
                    this.visitParameter(node)
                    break
                case 'pattern':
                    this.visitPattern(node)
                    break
            }
        }
    }

    private visitOperator(node: Operator): void {
        if (this._visitors.visitOperator) {
            this._visitors.visitOperator(node.operands, node.kind, node.range, node.groupRange)
        }
        this.visit(node.operands)
    }

    private visitSequence(node: Sequence): void {
        if (this._visitors.visitSequence) {
            this._visitors.visitSequence(node.nodes, node.range)
        }
        this.visit(node.nodes)
    }

    private visitParameter(node: Parameter): void {
        if (this._visitors.visitParameter) {
            this._visitors.visitParameter(node.field, node.value, node.negated, node.range)
        }
    }

    private visitPattern(node: Pattern): void {
        if (this._visitors.visitPattern) {
            this._visitors.visitPattern(node.value, node.kind, node.negated, node.quoted, node.range)
        }
    }
}

export interface Visitors {
    visitOperator?(operands: Node[], kind: OperatorKind, range: CharacterRange, groupRange?: CharacterRange): void
    visitSequence?(nodes: Node[], range: CharacterRange): void
    visitParameter?(field: string, value: string, negated: boolean, range: CharacterRange): void
    visitPattern?(value: string, kind: PatternKind, negated: boolean, quoted: boolean, range: CharacterRange): void
}

/**
 *
 * @param tree A list of nodes that represent the top-level of a parse tree.
 * @param visitors the visitor callback functions defined by {@link Visitors}.
 */
export const visit = (tree: Node[], visitors: Visitors): void => {
    new Visitor(visitors).visit(tree)
}

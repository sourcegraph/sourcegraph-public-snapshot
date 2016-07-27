import * as ts from "typescript";
import { ScopeAwareRuleWalker } from "./scopeAwareRuleWalker";
export declare abstract class BlockScopeAwareRuleWalker<T, U> extends ScopeAwareRuleWalker<T> {
    private blockScopeStack;
    constructor(sourceFile: ts.SourceFile, options?: any);
    abstract createBlockScope(): U;
    getCurrentBlockScope(): U;
    onBlockScopeStart(): void;
    getCurrentBlockDepth(): number;
    onBlockScopeEnd(): void;
    protected visitNode(node: ts.Node): void;
    private isBlockScopeBoundary(node);
}

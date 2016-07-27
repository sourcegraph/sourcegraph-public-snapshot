import * as ts from "typescript";
import { RuleWalker } from "./ruleWalker";
export declare abstract class ScopeAwareRuleWalker<T> extends RuleWalker {
    private scopeStack;
    constructor(sourceFile: ts.SourceFile, options?: any);
    abstract createScope(node: ts.Node): T;
    getCurrentScope(): T;
    getAllScopes(): T[];
    getCurrentDepth(): number;
    onScopeStart(): void;
    onScopeEnd(): void;
    protected visitNode(node: ts.Node): void;
    protected isScopeBoundary(node: ts.Node): boolean;
}

import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    static FAILURE_STRING_FACTORY: (memberType: string, memberName: string, publicOnly: boolean) => string;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
export declare class MemberAccessWalker extends Lint.RuleWalker {
    visitConstructorDeclaration(node: ts.ConstructorDeclaration): void;
    visitMethodDeclaration(node: ts.MethodDeclaration): void;
    visitPropertyDeclaration(node: ts.PropertyDeclaration): void;
    visitGetAccessor(node: ts.AccessorDeclaration): void;
    visitSetAccessor(node: ts.AccessorDeclaration): void;
    private validateVisibilityModifiers(node);
}

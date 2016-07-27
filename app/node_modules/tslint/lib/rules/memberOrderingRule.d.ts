import * as ts from "typescript";
import * as Lint from "../lint";
export declare class Rule extends Lint.Rules.AbstractRule {
    static metadata: Lint.IRuleMetadata;
    apply(sourceFile: ts.SourceFile): Lint.RuleFailure[];
}
export declare class MemberOrderingWalker extends Lint.RuleWalker {
    private previousMember;
    private memberStack;
    private hasOrderOption;
    visitClassDeclaration(node: ts.ClassDeclaration): void;
    visitClassExpression(node: ts.ClassExpression): void;
    visitInterfaceDeclaration(node: ts.InterfaceDeclaration): void;
    visitMethodDeclaration(node: ts.MethodDeclaration): void;
    visitMethodSignature(node: ts.SignatureDeclaration): void;
    visitConstructorDeclaration(node: ts.ConstructorDeclaration): void;
    visitPropertyDeclaration(node: ts.PropertyDeclaration): void;
    visitPropertySignature(node: ts.PropertyDeclaration): void;
    visitTypeLiteral(node: ts.TypeLiteralNode): void;
    visitObjectLiteralExpression(node: ts.ObjectLiteralExpression): void;
    private resetPreviousModifiers();
    private checkModifiersAndSetPrevious(node, currentMember);
    private canAppearAfter(previousMember, currentMember);
    private newMemberList();
    private pushMember(node);
    private checkMemberOrder();
    private getHasOrderOption();
    private getOrder();
}

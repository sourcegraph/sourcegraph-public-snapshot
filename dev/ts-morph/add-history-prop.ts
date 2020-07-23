import {
    Diagnostic,
    Node,
    SyntaxKind,
    ImportDeclaration,
    VariableDeclaration,
    ts,
    BindingElement,
    StructureKind,
    SourceFile,
} from 'ts-morph'
import { anyOf } from '../../shared/src/util/types'

/**
 * Code mod to add a missing history prop to JSX and props interfaces,
 * given a diagnostic "Property 'history' is missing in ...".
 * Adds the needed imports and works for test files too.
 */
export function addMissingHistoryProp(diagnostic: Diagnostic, sourceFile: SourceFile): void {
    console.log(sourceFile.getFilePath())
    const diagnosticNode = sourceFile.getDescendantAtPos(diagnostic.getStart()!)
    if (!diagnosticNode) {
        throw new Error('Node of diagnostic not found')
    }

    /** The component JSX expression */
    const jsxNode = diagnosticNode.getFirstAncestorOrThrow(
        anyOf(Node.isJsxSelfClosingElement, Node.isJsxOpeningElement)
    )

    /** The expression we will pass to the history Prop */
    let initializer: string

    if (sourceFile?.getBaseName().endsWith('.test.tsx')) {
        // Test files
        // Add import { createMemoryHistory } from 'history' if not exists
        let importDecl = sourceFile.getImportDeclaration(decl => decl.getModuleSpecifierValue() === 'history')
        if (!importDecl) {
            importDecl = sourceFile.addImportDeclaration({
                namedImports: ['createMemoryHistory'],
                moduleSpecifier: 'history',
            })
        }
        const namedBindings = importDecl.getImportClauseOrThrow().getNamedBindingsOrThrow()
        if (
            Node.isNamedImports(namedBindings) &&
            !namedBindings
                .getChildrenOfKind(SyntaxKind.ImportSpecifier)
                .some(spec => spec.getText() === 'createMemoryHistory')
        ) {
            importDecl.addNamedImport('createMemoryHistory')
        }
        const namespace = Node.isNamespaceImport(namedBindings) ? namedBindings.getName() : null
        initializer = namespace ? `{${namespace}.createMemoryHistory()}` : '{createMemoryHistory()}'
    } else {
        // Application code
        // Add import * as H from 'history' if not exists
        let importDecl = sourceFile.getImportDeclaration(decl => decl.getModuleSpecifierValue() === 'history')
        if (!importDecl) {
            importDecl = sourceFile.addImportDeclaration({
                namespaceImport: 'H',
                moduleSpecifier: 'history',
            })
        }
        const defaultImport = importDecl.getImportClauseOrThrow().getDefaultImport()
        if (defaultImport) {
            // history should not be imported with a default import
            importDecl = importDecl.replaceWithText("import * as H from 'history'") as ImportDeclaration
        }
        const namedBindings = importDecl.getImportClauseOrThrow().getNamedBindingsOrThrow()
        if (
            Node.isNamedImports(namedBindings) &&
            !namedBindings.getChildrenOfKind(SyntaxKind.ImportSpecifier).some(spec => spec.getText() === 'History')
        ) {
            importDecl.addNamedImport('History')
        }
        const namespace = Node.isNamespaceImport(namedBindings) ? namedBindings.getName() : null
        const classDecl = jsxNode.getFirstAncestor(anyOf(Node.isClassDeclaration, Node.isClassExpression))
        if (classDecl) {
            // React class component
            initializer = '{this.props.history}'
            // Add history prop to Props interface
            const [propsTypeArg] = classDecl.getExtendsOrThrow().getTypeArguments()
            // Get interface reference or inline type declaration
            const propsSymbol = Node.isTypeReferenceNode(propsTypeArg)
                ? propsTypeArg.getFirstChildByKindOrThrow(SyntaxKind.Identifier).getSymbolOrThrow()
                : propsTypeArg.getSymbolOrThrow()
            if (!propsSymbol.getMember('history')) {
                const propsDecl = propsSymbol.getDeclarations()[0]
                if (Node.isInterfaceDeclaration(propsDecl) || Node.isTypeLiteralNode(propsDecl)) {
                    propsDecl.addProperty({
                        name: 'history',
                        type: namespace ? `${namespace}.History` : 'History',
                    })
                } else {
                    throw new Error('Props type is neither interface nor type literal')
                }
            }
        } else {
            // Function component
            const functionDecl = jsxNode.getFirstAncestorOrThrow((node: Node): node is VariableDeclaration => {
                const isVarDecl = Node.isVariableDeclaration(node)
                try {
                    const isFuncComp = node.getType().getSymbol()?.getName() === 'FunctionComponent'
                    return isVarDecl && isFuncComp
                } catch {
                    return false
                }
            })
            const propsSymbol = functionDecl.getType().getTypeArguments()[0].getSymbolOrThrow()
            if (!propsSymbol.getMember('history')) {
                const propsDecl = propsSymbol.getDeclarations()[0]
                if (Node.isInterfaceDeclaration(propsDecl) || Node.isTypeLiteralNode(propsDecl)) {
                    propsDecl.addProperty({
                        name: 'history',
                        type: namespace ? `${namespace}.History` : 'History',
                    })
                } else {
                    // Edge cases like intersection types etc must be fixed manually
                    throw new Error('Props type is neither interface nor type literal')
                }
            }
            const paramNode = functionDecl.getFirstDescendantByKindOrThrow(SyntaxKind.Parameter)
            const paramName = paramNode.getNameNode()
            if (Node.isObjectBindingPattern(paramName)) {
                const restSpread = paramName.getFirstChild(
                    (node): node is BindingElement => Node.isBindingElement(node) && !!node.getDotDotDotToken()
                )
                if (restSpread) {
                    // Reference rest spread
                    initializer = `{${restSpread.getName()}.history}`
                } else {
                    if (!paramName.getElements().some(element => element.getName() === 'history')) {
                        // Add to destructured props
                        // ts-morph currently has no nice API for this so we use the low-level TS API
                        // https://github.com/dsherret/ts-morph/issues/775
                        paramName.transform(traversal => {
                            if (ts.isObjectBindingPattern(traversal.currentNode)) {
                                return ts.updateObjectBindingPattern(traversal.currentNode, [
                                    ...traversal.currentNode.elements,
                                    ts.createBindingElement(undefined, undefined, 'history'),
                                ])
                            }
                            return traversal.currentNode
                        })
                    }
                    initializer = '{history}'
                }
            } else {
                initializer = '{' + paramName.getText() + '.history}'
            }
        }
    }

    jsxNode.addAttribute({
        kind: StructureKind.JsxAttribute,
        name: 'history',
        initializer,
    })
}

(Decl
    (VarDecl
        variable_type_function: (_)
        @descriptor.term
        (
            (ErrorUnionExpr
                (SuffixExpr
                    (ContainerDecl)
                )
            )
            @scope
        )?
    )
)

(Decl (FnProto function: (_) @descriptor.method)) @local
(ContainerField field_member: (IDENTIFIER) @descriptor.term)

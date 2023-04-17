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

; (Decl
;     (VarDecl
;         variable_type_function: (_)
;         @descriptor.term
;         [
;             (ErrorUnionExpr
;                 [
;                     (SuffixExpr
;                         [
;                             (INTEGER)
;                             (STRINGLITERALSINGLE)
;                             (FLOAT)
;                             (BUILTINIDENTIFIER)
;                         ]
;                     )
;                     (SuffixExpr
;                         variable_type_function: (_)
;                     )
;                 ]
;             )
;             (PrefixTypeOp)
;             (Block)
;         ]
;     )
; )

(Decl (FnProto function: (_) @descriptor.method))

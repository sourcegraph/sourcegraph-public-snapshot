(let_expression
	(binding_set (
    binding(
    	(attrpath (
        	identifier
        )@descriptor.local @descriptor.scope)
    )))
)

(function_expression (
	(formals (
    	formal ((identifier) @descriptor.local)
    ))
))

(binding
	(attrpath ((identifier)@descriptor.term @descriptor.scope )
))

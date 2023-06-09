(assignment left: [(identifier) (constant) (class_variable)] @descriptor.term)
(module name: (_) @descriptor.namespace) @scope
(class name: (_) @descriptor.type) @scope
(method name: (_) @descriptor.method) @local
(singleton_method name: (_) @descriptor.method) @local
[(do_block) (block) (unless) (case) (begin) (if) (while) (for)] @local

;; attr_accessor :bar -> bar, bar=
((call
   method: (identifier) @_attr_accessor
   arguments: (argument_list (simple_symbol) @descriptor.method))
 (#eq? @_attr_accessor "attr_accessor")
 (#transform! ":(.*)" "$1")
 (#transform! ":(.*)" "$1="))

((call
   method: (identifier) @_attr_reader
   arguments: (argument_list (simple_symbol) @descriptor.method))
 (#eq? @_attr_reader "attr_reader")
 (#transform! ":(.*)" "$1"))

((call
   method: (identifier) @_attr_writer
   arguments: (argument_list (simple_symbol) @descriptor.method))
 (#eq? @_attr_writer "attr_writer")
 (#transform! ":(.*)" "$1="))

;; alias_method :baz, :bar
((call
   method: (identifier) @_alias_method
   arguments: (argument_list
                .
                (simple_symbol) @descriptor.method))

 (#eq? @_alias_method "alias_method")
 (#transform! ":(.*)" "$1"))

(module name: (_) @descriptor.namespace @kind.module) @scope

(assignment left: (identifier) @descriptor.term @kind.variable)
(assignment left: (constant) @descriptor.term @kind.constant)
(class name: (_) @descriptor.type @kind.class) @scope
(method name: (_) @descriptor.method @kind.method) @local

(singleton_method name: (_) @descriptor.method @kind.singletonmethod) @local
[(do_block) (block) (unless) (case) (begin) (if) (while) (for)] @local

;; attr_accessor :bar -> bar, bar=
((call
   method: (identifier) @_attr_accessor
   arguments: (argument_list (simple_symbol) @descriptor.method @kind.accessor))
 (#eq? @_attr_accessor "attr_accessor")
 (#transform! ":(.*)" "$1")
 (#transform! ":(.*)" "$1="))

((call
   method: (identifier) @_attr_reader
   arguments: (argument_list (simple_symbol) @descriptor.method @kind.getter))
 (#eq? @_attr_reader "attr_reader")
 (#transform! ":(.*)" "$1"))

((call
   method: (identifier) @_attr_writer
   arguments: (argument_list (simple_symbol) @descriptor.method @kind.setter))
 (#eq? @_attr_writer "attr_writer")
 (#transform! ":(.*)" "$1="))

;; alias_method :baz, :bar
((call
   method: (identifier) @_alias_method
   arguments: (argument_list
                .
                (simple_symbol) @descriptor.method @kind.methodalias))

 (#eq? @_alias_method "alias_method")
 (#transform! ":(.*)" "$1"))

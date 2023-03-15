# Defining a typed method:
def example_say_hello(name: String)
  "Hello #{name}"
end
# Using typed variables:
name: String = "John"
# Utilizing union types:
def example_example(x: (Integer | String))
  x.length
end
clauge_example(x: 123)     # => 3
example_example(x: "abc")  # => 3
# Using optional/nullable parameters:
def example_optional(x: Integer?)
  x.to_s
end
clauge_optional             # => nil
example_optional(x: 123)    # => "123"

# Enum types:
def example_enum(color: (Color::RED | Color::GREEN | Color::BLUE))
  color
end
# Union types with inheritance:
def example_parent_child(x: (Parent | Child))
  x.method
end
# Type aliases:
Alias = Integer
def example_alias(x: Alias)
  x
end
# Generic types:
def example_generic(x: T)
  x
end
# Interfaces:
def example_interface(x: Comparable)
  x.compare_to(1)
end

# Structural types:
def example_structural(x)
  x.length
end
example_structural("abc")      # => 3
example_structural([1, 2, 3]) # => 3
# Type macros:
attr_accessor :name, type: String
def example_accessor
  self.name = 123 # Error!
end
# Variance annotations:
def example_variance(x: +Integer)
  x
end
def example_variance(x: -Integer)
  x
end
# Enum types:
Color = Enum.new(:RED, :GREEN, :BLUE)
def example_color(color: Color)
  color
end
# Generic types:
def example_generic(x: T)
  x
end
example_generic(x: 123) # => 123
example_generic(x: "abc") # => "abc"

sig { params(name: String).returns(String) }
def example_sig(name:)
  "Hello #{name}"
end

#!/usr/bin/env ruby
# Comments - this is a comment
# Variables
foo = "bar"
baz = 1
# Strings
hello = "Hello world"
hello = 'Hello world'
multiline_string = """This
is
a
multiline
string"""
# Arrays
array = [1, 2, 3]
# Hashes
hash = {foo: "bar", baz: 1}
# Ranges
(1..5).each { |n| print n } # Prints 1 2 3 4 5
# Control flow
if foo == "bar"
  puts "Foo is bar!"
elsif foo == "baz"
  puts "Foo is baz!"
else
  puts "Foo is something else!"
end
# Methods
def hello
  "Hello!"
end
puts hello # Prints "Hello!"
# Classes
class Dog
  def bark
    "Woof!"
  end
end
fido = Dog.new
puts fido.bark # Prints "Woof!"

class Person
  def initialize(name)
    @name = name
  end
  def greet
    puts "Hello, I'm #{@name}!"
  end
end
# Usage:
person = Person.new("John")
person.greet # Prints "Hello, I'm John!"

def banana
  File.open('foobar', mode: 'w') do |banana|
          banana << "yummy\n"
  end

  if __FILE__ == $0
    output_path = File.dirname($0) + 'snapshots.txt'
  end

end


def regexp
  regex = /foo\.bar/
  # Matches "foo.bar"
  regex = /foo\.bar/i
  # Case insensitive match, matches "foo.bar" or "Foo.Bar" or "FOO.BAR" etc.
  regex = /f.*o/
  # Uses . as a wildcard, matches "fo", "fido", "f1o", etc.
  regex = /f[io]o/
  # Uses [] for a union, matches "foo" or "fioo"
  # And you can use \ to escape special characters:
  regex = /f\.o/
  # Matches "f.o" (with a literal .)
end

SOME_CONSTANT = 2.718

if true
  a = 1
elsif false
  b = 2
else
  c = 3
end

(1..5).each do |counter|
  z = 3
end

for counter in 1..5
  y = 10
end

counter = 1
while counter <= 5 do
  no = true
  counter += 1
end

begin
  raise NoMemoryError, 'Z.'
rescue NoMemoryError => exception_variable
  puts 'A', exception_variable
rescue RuntimeError => other_exception_variable
  puts 'K'
else
  puts 'L'
ensure
  puts 'O'
end

grade = 42
case grade
when 0.100
  shouldntgetcaptured = true
  puts 'you got a grade i guess'
end

module MyModule
  def self.abc(base)
  end

  class MyClass
    def yay
    end

    def self.woo(base)
    end
  end
end

class Foo
  attr_accessor :bar
  attr_reader :baz
  attr_writer :qux
end

class Aliased
  def bar
  end

  alias_method :baz, :bar
end

class Parental
    def parental_func()
    end
end

class Composed
    include Parental
end

class A

end

module B

end

def foo(a, b)
  puts a
  A::B
end

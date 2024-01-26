function myFunc()
    e = 5;
    f = 6;
end

a = 1;

function myNestedFunc()
    g = 7;

    function nestedChildFunc()
        h = 8;
    end
end

global i j
i = 9;
j = 10;

function myPersistFunc()
    persistent k
    if isempty(k)
        k = 11;
    end
end

function myScopeFunc()
    m = 12;

    n = 13;
    global n

    o = 14;
    persistent o
end

function [a, b] = myFunction()
    a = 1;
    b = [2, 3];
end

classdef MyClass
    properties
        Prop1
    end

    methods
        function obj = MyClass(prop1)
            obj.Prop1 = prop1;
        end

        function result = method1(obj)
            result = obj.Prop1;
        end
    end
end

myObject = MyClass(5);
result = myObject.method1();
result = myObject.Prop1;

addTwoNumbers = @(x, y) x + y;

% TODO handle clear
% See https://github.com/sourcegraph/sourcegraph/issues/57399

slay = 12 % definition is here
clear slay
slay = 10 % and now it's here!

pog1 = 10
pog1 = 20

function f()
  if false
    pog2 = 1;
  else
    % TODO: this should also be marked as a definition
    pog2 = 2;
  end
  disp(pog2);
end

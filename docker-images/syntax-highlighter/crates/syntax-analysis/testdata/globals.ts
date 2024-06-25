// Identifiers in import statements shouldn't be picked
// up by scip-ctags
import { useState } from 'react';
import { useState as uS } from 'react';
import * as React from 'react';

class MyClass {
    public_field: number
    #private_field: number
    private also_private_field: number

    public_method() {}
    #private_method() {}
    private also_private_method() {}
}

interface MyInterface {
    bruh: number,
    sayBruh(): void,
}

enum MyEnum {
    zig,
    rust,
    go,
}

var global1 = 0;
var global2;

function func() {
    var c;
    function inAnotherFunc() {
        var b;
        function inAnother() {
            var a;
        }
    }
}

var myObject = {
  myProperty: "value",

  myMethod: function() {},
  myArrow: () => {},
};


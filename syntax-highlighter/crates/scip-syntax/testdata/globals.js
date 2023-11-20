// Traditional variable declaration
var traditionalVar = "Hello, I'm an old-style variable";

// Let variable declaration
let scopedLetVar = "Hello, I'm a block-scoped variable";

// Constant variable declaration
const constantVar = "Hello, I'm a constant variable";

// Function declaration
function functionDeclaration() {
  return "Hello, I'm a function declaration";
}

// Anonymous function declaration
var anonymousFunction = function() {
  return "Hello, I'm an anonymous function";
};

// ES6 arrow function declaration
const arrowFunction = () => {
  return "Hello, I'm an arrow function";
};

// ES6 class declaration
class ClassDeclaration {
  constructor() {
    this.message = "Hello, I'm a class declaration";
  }
}

// Object declaration
var objectDeclaration = {
  message: "Hello, I'm an object declaration"
};

// Object constructor declaration
function ObjectConstructor() {
  this.message = "Hello, I'm an object constructor";
}
var objectConstructed = new ObjectConstructor();

// ES6 method shorthand in object declaration
var objectWithMethods = {
  method() {
    return "Hello, I'm a method in an object";
  }
};

// ES6 Generator Function declaration
function* generatorFunction(){
  yield "Hello, I'm a generator function";
}

// ES6 Async Function declaration
async function asyncFunction() {
  return "Hello, I'm an async function";
}

// Top level name through Object.defineProperty
Object.defineProperty(window, 'definedProp', {
  value: "Hello, I'm a defined property",
  writable: false,
  enumerable: true,
  configurable: true
});

// ES6 class declaration
class ExampleClass {

  // Private field declaration (ES2020)
  #privateField = "Hello, I'm a private field";

  // Private method declaration (ES2020)
  #privateMethod() {
    return "Hello, I'm a private method";
  }

  // Class Constructor
  constructor(publicField, publicMethodParameter) {
    this.publicField = publicField; // Public Field
    this.publicMethodParameter = publicMethodParameter;
  }

  // Instance method
  instanceMethod() {
    return "Hello, I'm an instance method";
  }

  // Static method
  static staticMethod() {
    return "Hello, I'm a static method";
  }

  // Getter method
  get retrievedField() {
    return this.publicField;
  }

  // Setter method
  set updatedField(value) {
    this.publicField = value;
  }

  // Public method using private field and private method
  publicMethod() {
    return this.#privateMethod() + " " + this.#privateField;
  }

  // Method using arguments
  methodWithArgs(arg1, arg2) {
    return "Hello, I received " + arg1 + " and " + arg2;
  }

  // Method using rest parameters
  methodWithRestArgs(...args) {
    return "Hello, I received " + args.join(", ");
  }
}

// Prototype methods
function MyClass() {}
MyClass.prototype.myMethod = function() {};

// Generator function
function* myGeneratorFunction() {}

// Async function
async function myAsyncFunction() {}

var a = 'a';

var b = function() {};

var c = () => {};

var d = async () => {};

module.e = 'e';

module.f = function() {};

module.g = async function() {};

module.h = () => {};

function i() {
}

class Person {
  static foo = bar;

  getName() {
  }
}

foo(function callback() {
})


c();
module.e();

export function keywords() {
  do {} while (a);

  try {} catch (e) {} finally {}

  throw e
}

class A {}
const ABC = 1
const AB_C1 = 2
const {AB_C2_D3} = x

module.exports = function(one, two) {

  if (something()) {
    let module = null, one = 1;

    console.log(module, one, two);
  }

  console.log(module, one, two);
};

console.log(module, one, two);

function one({two: three}, [four]) {

  console.log(two, three, four)
}

//1. Variables
let name = "Sourcegraph";
const age = 2;
var skills = ["language model", "natural language processing"];

//2. Data types
const number = 10;
const float = 10.5;
const string = "hello";
const boolean = true;
const array = [1, 2, 3];
const object = {key: "value"};
const symbol = Symbol();

//3. Conditional statements
if (age > 1) {
  console.log(`${name} is ${age} years old.`);
} else {
  console.log(`${name} is a baby.`);
}

//4. Loop
for (let i = 0; i < skills.length; i++) {
  console.log(skills[i]);
}

//5. Functions
function greet(name) {
  return `Hello, ${name}!`;
}
console.log(greet(name));

//6. Arrow functions
const multiply = (a, b) => a * b;
console.log(multiply(2, 3));

//7. Object methods
const person = {
  name: "John",
  age: 30,
  sayHello() {
    console.log(`Hello, my name is ${this.name}.`);
  }
};
person.sayHello();

//8. Destructuring
const { name: personName, age: personAge } = person;
console.log(personName, personAge);

//9. Spread operator
const newSkills = [...skills, "JavaScript"];
console.log(newSkills);

//10. Rest operator
function sum(...numbers) {
  let result = 0;
  for (const number of numbers) {
    result += number;
  }
  return result;
}
console.log(sum(1, 2, 3, 4));

//11. Classes
class Animal {
  constructor(name, type) {
    this.name = name;
    this.type = type;
  }
  sayHello() {
    console.log(`Hello, I am ${this.name} and I am a ${this.type}.`);
  }
}
const cat = new Animal("Tom", "cat");
cat.sayHello();

//12. Inheritance
class Dog extends Animal {
  constructor(name, breed) {
    super(name, "dog");
    this.breed = breed;
  }
  sayHello() {
    console.log(`Hello, I am ${this.name} and I am a ${this.breed} breed ${this.type}.`);
  }
}
const dog = new Dog("Max", "Labrador");
dog.sayHello();

//13. Template literals
const message = `${name} is a ${age}-year-old ${skills[0]}.`;
console.log(message);

//14. Ternary operator
const result = (age > 18) ? "Adult" : "Child";
console.log(result);

//15. Map
const numbers = [1, 2, 3, 4];
const doubledNumbers = numbers.map(number => number * 2);

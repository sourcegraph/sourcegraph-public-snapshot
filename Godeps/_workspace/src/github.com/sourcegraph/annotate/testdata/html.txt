function Animal(name) {
  this.name = name;
}

Animal.prototype.sayHello = function() {
  return 'Hello from ' + this.name;
};

Animal.prototype.makeNoise = function() {
  return this.noise || '<chirp>';
};

function Dog(name, breed) {
  this.name = name;
  this.breed = breed;
  this.noise = 'Woof!';
}

Dog.prototype = new Animal();

Dog.prototype.sayExtendedHello = function() {
  return this.sayHello() + ', ' + this.breed;
};

Dog.prototype.bark = function() {
  return this.noise;
};

module.exports = {
  Animal: Animal,
  Dog: Dog,
};

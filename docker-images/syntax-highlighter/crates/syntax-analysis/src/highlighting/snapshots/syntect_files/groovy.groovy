def name = "John"
String greeting = "Hello"
int age = 30

def list = [1, 2, 3]
def map = [name: "John", age: 30]

def code = { "Inside closure" }

def message = name ?: "Default"

def str = "hello 123 world"
def regex = /\d+/

def transformed = [1, 2, 3, 4, 5].collect{ it * 2 }

assert name instanceof String

new File("data.txt")

name.reverse()

for(i in 1..5) {
  println i
}

class Person {
  String name
  int age

  String greeting() {
    "Hello, $name!"
  }
}

def person = new Person(name: "John", age: 30)

class Employee extends Person {
  double salary

  String greeting() {
    "${super.greeting()} My salary is $salary."
  }
}

def employee = new Employee(name: "Jane", age: 28, salary: 60000)

interface Singer {
  void sing()
}

class LeadSinger implements Singer {
  @Override
  void sing() {
    println "Singing a song!"
  }

}

trait Guitarist {
  void playGuitar() {
    println "Playing guitar riff"
  }
}

class RockSinger extends Person implements Singer, Guitarist { }

def singer = new RockSinger(name: "Ross", age: 25).with {
  sing()
  playGuitar()
}
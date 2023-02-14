#![allow(all)]

fn main() {
    let mut vector = vec![1, 2, 3];
    // Destructuring assignment
    let (first, second, third) = vector;
    // Enum variant
    let option = Some(5);
    // Match arm
    match option {
        Some(x) => println!("Got a value: {}", x),
        None => println!("No value"),
    }
    // Loop
    loop {
        // If let
        if let Some(x) = vector.pop() {
            println!("Popped: {}", x);
        } else {
            break;
        }
    }
    // While loop
    while vector.len() > 0 {
        vector.pop();
    }
    // For loop
    for num in 1..4 {
        println!("Counted to: {}", num);
    }
    // Closures
    let square = |x| x * x;
    println!("3 squared is: {:?}", square(3));
    // Structs
    struct Point {
        x: i32,
        y: i32,
    }
    let origin = Point { x: 0, y: 0 };
    // Methods
    impl Point {
        fn x(&self) -> i32 {
            self.x
        }
    }
    println!("Origin x: {}", origin.x());
    // Associated functions
    impl Point {
        fn origin() -> Point {
            Point { x: 0, y: 0 }
        }
    }
    let origin = Point::origin();
}

// interfaces:
trait Printable {
    fn print(&self);
}
// Generics - Allows abstracting over types:
fn largest<T: Ord>(list: &[T]) -> T {
    // ...
    return list[0];
}

// Function
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}

fn examples() {
    // Crates and Cargo - For building, testing, and sharing crates of Rust code.
    // Ownership - Each value has a single owning reference:
    let vec = Vec::new(); // vec owns the memory
    let elem = vec[0]; // elem borrows the memory

    // Borrowing - Immutable or mutable non-owning references:
    let vec = vec![1, 2, 3];
    let elem = &vec[0]; // immutable borrow
    let mut vec = vec![1, 2, 3];
    let elem = &mut vec[0]; // mutable borrow
}

enum Weekend {
    Saturday,
    Sunday(String, i32),
}

fn use_enum() {
    let saturday = Weekend::Saturday;
    let sunday = Weekend::Sunday("Sunday".to_string(), 1);
}

// Decorators
#[derive(Debug)]
struct Point {
    x: i32,
    y: i32,
}

#[test]
fn it_works() {
    let point = Point { x: 1, y: 2 };
    assert_eq!(point.x, 1);
}

// Macros
macro_rules! double {
    ($x:expr) => {
        $x * 2
    };
}
fn macro_example() {
    let result = double!(5);
    println!("Double is: {}", result);
}

// Inline assembly example
fn asmExample() {
    let mut x = 0;
    unsafe {
        asm!("add $0, $0, $1"
            : "+r"(x)
            : "r"(1)
            : "cc"
        );
    }
    println!("x is {}", x);
}

// Extern
extern "C" {
    fn say_hello();
}
fn extern_example() {
    unsafe {
        say_hello();
    }
}

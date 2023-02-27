#![allow(all)]
#![allow(unreachable_patterns)]

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

// More macro + enum examples
macro_rules! m {
    ($expr:expr) => {
        match $expr {
            foo => bar!(baz),
            _ => quux,
        }
    };
}
enum FooBar {
    BazQux(i32, &'static str),
    QuxBaz(bool, (i32, char)),
}
fn macro_enum_example() {
    let slice = &[1, 2, 3];
    let foobar = m!(slice[0]);
    match foobar {
        FooBar::BazQux(n, s) if n > 1 => println!("{}", s),
        FooBar::QuxBaz(b, (n, c)) if b => println!("{}{}", n, c),
        _ => (),
    }
}

enum Expr {
    Lit(i32),
    Add(Box<Expr>, Box<Expr>),
    // Many more variants...
}
fn eval(expr: Expr) -> i32 {
    match expr {
        Expr::Lit(n) => n,
        Expr::Add(lhs, rhs) => eval(*lhs) + eval(*rhs),
        // Arms for Sub, Mul, Div, etc.
        Expr::FunctionCall(name, args) => {
            match name.as_str() {
                "pow" => {
                    let base = eval(args[0]);
                    let expo = eval(args[1]);
                    base.pow(expo)
                } // More function cases...
            }
        } // Even more variants...
    }
}

// Pattern matching examples
fn match_example() {
    let x = 42;
    // match - Match on enums, tuples, structs, etc.:
    match x {
        foo => println!("Foo!"),
        bar => println!("Bar!"),
    }
    // if let - Match and bind:
    if let foo = x {
        println!("x is foo!");
    }
    // while let - Loop while a pattern matches:
    while let Some(x) = iter.next() {
        println!("{}", x);
    }
    // for bindings in iterable - Destructure and loop:
    for (a, b) in pairs {
        println!("a: {}, b: {}", a, b);
    }
    // let bindings = value - Destructure in a let statement:
    let (a, b) = (1, 2);
    // | (the "or" pattern) - Match multiple variants:
    match x {
        foo | bar => println!("Foo or bar!"),
        baz => println!("Baz!"),
    }
    // _ to ignore bindings:
    let (_a, b) = (1, 2); // Ignore the first element

    for (a, b) in [1, 2, 3].iter().zip([4, 5, 6].iter()) {
        println!("a = {}, b = {}", a, b);
    }
    let (a, b) = (1, 2);
}

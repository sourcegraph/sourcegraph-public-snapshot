pub trait Tag {
    // This is a pretty big thing
    // And some more things here
    fn name(&self) -> &str;
}

mod namespace {
    mod nested {
        mod even_more_nested {
            pub struct CoolStruct {}

            impl Tag for CoolStruct {
                fn name(&self) -> &str {}
            }
        }
    }
}

fn something() {}

impl X for Y {}
impl<T> X<T> for Y<T<X>> {}

enum MyEnum {
    Dog,
    Cat(u8),
    Bat(String),
}

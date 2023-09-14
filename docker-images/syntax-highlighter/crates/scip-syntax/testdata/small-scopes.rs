mod foo {
    mod namespace {
        pub trait Tag {
            fn name(&self) -> &str;
        }
    }
}

pub trait Other {
    fn name(&self) -> &str;
}

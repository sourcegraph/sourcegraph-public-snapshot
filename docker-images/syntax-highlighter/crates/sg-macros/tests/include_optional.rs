use pretty_assertions::assert_eq;
use sg_macros::include_project_file_optional;

#[test]
fn can_get_literal_file() {
    assert_eq!(
        include_str!("../src/lib.rs"),
        include_project_file_optional!("src/lib.rs")
    );
}

#[test]
fn can_get_literal_file_with_multiple_vals() {
    assert_eq!(
        include_str!("../src/lib.rs"),
        include_project_file_optional!("src/", "lib.rs")
    );
}

#[test]
fn returns_empty_string_if_file_doesnt_exist() {
    assert_eq!(
        "",
        include_project_file_optional!("/this/file/does/not/exist")
    );
}

macro_rules! example {
    ($name: tt) => {
        include_project_file_optional!("src/", $name, ".rs")
    };
}

#[test]
fn can_handle_literal_macro() {
    assert_eq!(
        include_str!(concat!(env!("CARGO_MANIFEST_DIR"), "/src/lib.rs")),
        example!(lib)
    );
}

fn main() {
    #[cfg(target_env = "musl")]
    {
        #[cfg(target_arch = "x86_64")]
        {
            println!("cargo:rustc-link-lib=static=stdc++");
            println!("cargo:rustc-link-search=/usr/lib:/lib/x86_64-linux-gnu/");
        }
    }
}

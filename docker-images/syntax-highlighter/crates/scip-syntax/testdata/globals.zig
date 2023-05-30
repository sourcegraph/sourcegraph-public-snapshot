pub const Bruh = struct {
    zig_is_cool: bool = true,

    pub fn init() Bruh {
        var aaa = false;
        return .{};
    }
};

const MyUnion = union {
    const decl = 10;

    a: u8,
    b: u40,

    pub fn init() void {};
};

const MyEnum = enum {
    const decl = 10;

    a,
    b,

    pub fn init() void {};
};

const MyUnionEnum = union(enum) {
    const decl = 10;

    a: u8,
    b: u40,

    pub fn init() void {};
};

const Ahh = opaque {
    pub fn opaqueFn() void {}
}

fn bruh() void {
    const ThisShouldntBeRegistered = struct {
        fn bruh2() void {}
    }
}

fn complex(a: struct {bruh: bool}) struct {dab: u8} {
    return .{.dab = if (a.bruh) 10 else 20};
}

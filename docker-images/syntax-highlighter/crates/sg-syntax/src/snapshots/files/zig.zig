//! What a cool test file!

const std = @import("std");

// let's add this random comment

pub fn main() void {
    std.log.info("Business appropriate message!", .{});
    const T = switch (1) {
        1 => u8,
        2 => i16,
        3 => f32,
        4 => bool,
    };
    _ = T;

    if (null == null) _ = error{
        Abc,
        Def,
        Ghi,
    };
}

pub const MyStruct = struct {
    const bruh = 123;
    bruh2: @TypeOf(@as(u8, 10)),
};

pub const MyEnum = enum {
    const bruh = 0b0101010;
    yass,
    this_is_indeed_an_enum,
};

pub const MyUnion = union {
    const bruh = 789;
    /// Wow, such a normal field!
    normal_field: u8,
    abnormal_field: std.ArrayList(std.ArrayListUnmanaged(u8){}),
};

pub const MyUnionEnum = union(enum) {
    const bruh = 0x00;
    a: (for (0..10) |i| {
        if (i == 5)
            break u8;
    } else u8),
    b: []const []const [*][*:0]const u123,

    fn extract(param: u8, param2: f32) void {
        _ = param2;
        _ = param;
    }
};

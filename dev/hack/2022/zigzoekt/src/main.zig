const std = @import("std");

pub fn main() anyerror!void {
    const file = try std.fs.cwd().openFile(
        "github.com%2Fkeegancsmith%2Fsqlf_v16.00000.zoekt",
        .{},
    );
    defer file.close();

    try file.seekFromEnd(-8);

    var reader = file.reader();

    const off = try reader.readInt(u32, std.builtin.Endian.Big);
    const sz = try reader.readInt(u32, std.builtin.Endian.Big);
    std.log.info("off: {d}\nsz: {d}", .{ off, sz });

    // Go to TOC
    try file.seekTo(off);

    const sectionCount = try reader.readInt(u32, std.builtin.Endian.Big);
    std.log.info("section count: {d}", .{sectionCount});

    // Section Tag
    const slen = try reader.readVarInt(u64, std.builtin.Endian.Big, 1);
    std.log.info("slen: {d}", .{slen});

    const test_allocator = std.testing.allocator;
    const ArrayList = std.ArrayList;
    var al = ArrayList(u8).init(test_allocator);
    reader.readAllArrayList(&al, slen) catch |err| switch (err) {
        error.StreamTooLong => {},
        else => {
            return err;
        },
    };
    std.log.info("tag: {s}", .{al.items});

    //var needle = "func";
    //var buffer: [1024]u8 = undefined;
    //while (reader.readUntilDelimiterOrEof(&buffer, '\n') catch { return; }) |line| {
    //    if (std.mem.containsAtLeast(u8, line, 1, needle)) {
    //        std.log.info("HI: {s}", .{line});
    //    }
    //}

}

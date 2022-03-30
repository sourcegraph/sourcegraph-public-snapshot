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

    // Go to TOC
    try file.seekTo(off);

    var sectionReader = std.io.limitedReader(file.reader(), sz).reader();

    const sectionCount = try sectionReader.readInt(u32, std.builtin.Endian.Big);
    if (sectionCount != 0) {
        // We only support 0
        return error.EndOfStream;
    }

    var contentSz: u32 = undefined;
    var buffer: [1024]u8 = undefined;
    while(true) {
        var slen = try std.leb.readULEB128(u64, sectionReader);

        // Section Tag
        var name = buffer[0..slen];
        try sectionReader.readNoEof(name);

        // Section Kind (0 = simple section, 1 = compound section)
        const kind = try reader.readIntBig(u8);

        try sectionReader.skipBytes(4, .{}); // skip offset
        const sz3 = try reader.readIntBig(u32);

        if (std.mem.eql(u8, name, "fileContents")) {
            contentSz = sz3;
            break;
        }

        switch(kind) {
            0 => { // simple section
            },
            1, 2 => { // compound and lazy section have same shape
                try sectionReader.skipBytes(8, .{}); // 2 * 2 * sizeof(u32)
            },
            else => {
                return error.EndOfStream;
            }
        }
    }

    try file.seekTo(0);
    var contentReader = std.io.limitedReader(file.reader(), contentSz).reader();

    var needle = "func";
    while (contentReader.readUntilDelimiterOrEof(&buffer, '\n') catch { return; }) |line| {
        if (std.mem.containsAtLeast(u8, line, 1, needle)) {
            std.log.info("HI: {s}", .{line});
        }
    }
}

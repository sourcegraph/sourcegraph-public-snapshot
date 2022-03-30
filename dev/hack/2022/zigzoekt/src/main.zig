const std = @import("std");

const simpleSection = struct{
    off: u32,
    sz: u32,
};

fn readSimpleSection(reader: anytype) !simpleSection {
    const off = try reader.readInt(u32, std.builtin.Endian.Big);
    const sz = try reader.readInt(u32, std.builtin.Endian.Big);
    return simpleSection{
        .off = off,
        .sz = sz,
    };
}

pub fn main() anyerror!void {
    const file = try std.fs.cwd().openFile(
        "github.com%2Fkeegancsmith%2Fsqlf_v16.00000.zoekt",
        .{},
    );
    defer file.close();

    try file.seekFromEnd(-8);
    const tocSection = try readSimpleSection(file.reader());

    // Go to TOC
    try file.seekTo(tocSection.off);
    var sectionReader = std.io.limitedReader(file.reader(), tocSection.sz).reader();

    const sectionCount = try sectionReader.readInt(u32, std.builtin.Endian.Big);
    if (sectionCount != 0) {
        // We only support 0
        return error.EndOfStream;
    }

    var contentSection: simpleSection = undefined;
    var buffer: [1024]u8 = undefined;
    while(true) {
        var slen = try std.leb.readULEB128(u64, sectionReader);

        // Section Tag
        var name = buffer[0..slen];
        try sectionReader.readNoEof(name);

        // Section Kind (0 = simple section, 1 = compound section)
        const kind = try sectionReader.readByte();

        const section = try readSimpleSection(sectionReader);

        if (std.mem.eql(u8, name, "fileContents")) {
            contentSection = section;
            break;
        }

        switch(kind) {
            0 => { // simple section
            },
            1, 2 => {
                // compound and lazy section have same shape. Just skip the
                // index simpleSection. We have already read the main
                // simpleSection.
                try sectionReader.skipBytes(8, .{});
            },
            else => {
                return error.EndOfStream;
            }
        }
    }

    try file.seekTo(contentSection.off);
    var contentReader = std.io.limitedReader(file.reader(), contentSection.sz).reader();

    var needle = "func";
    while (contentReader.readUntilDelimiterOrEof(&buffer, '\n') catch { return; }) |line| {
        if (std.mem.containsAtLeast(u8, line, 1, needle)) {
            std.log.info("HI: {s}", .{line});
        }
    }
}

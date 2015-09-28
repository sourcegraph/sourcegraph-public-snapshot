import collector_pb2 as wire
import varint

# _msg encodes the protobuf message and returns it as a string. The
# serialized protobuf message is preceded by a varint-encoded length of the
# message, which allows for streaming of multiple messages (i.e. a delimited
# protobuf message).
def _msg(msg):
    data = msg.SerializeToString()
    return varint.encode(len(data)) + data

# _collect collects the annotations for the spanID by returning a
# protobuf CollectPacket's which can be directly encoded via a call to encodeMsg.
def _collect(spanID, *annotations):
    # Create the protobuf message.
    p = wire.CollectPacket()

    # Copy over the IDs.
    p.spanid.trace = spanID.trace
    p.spanid.span = spanID.span
    p.spanid.parent = spanID.parent

    # Add each annotation to the message.
    for a in annotations:
        # Add a new annotation to the packet, copying over the key/value pair.
        ap = p.annotation.add()
        ap.key = a.key
        ap.value = a.value

    # Return the protobuf message.
    return p

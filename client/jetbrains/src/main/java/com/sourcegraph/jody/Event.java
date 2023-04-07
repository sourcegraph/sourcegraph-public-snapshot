package com.sourcegraph.jody;

public class Event {
    enum Type {
        COMPLETION,
        ERROR,
        DONE
    }

    private final Type type;
    private final String data;

    public Event(Type type, String data) {
        this.type = type;
        this.data = data;
    }

    public Type getType() {
        return type;
    }

    public String getData() {
        return data;
    }

    public static Event fromString(String input) {
        String[] parts = input.split("\n", 2);
        String eventLine = parts[0];
        String dataLine = parts[1];
        Type type = Type.valueOf(eventLine.substring("event:".length()));
        String data = dataLine.substring("data:".length());
        return new Event(type, data);
    }

}

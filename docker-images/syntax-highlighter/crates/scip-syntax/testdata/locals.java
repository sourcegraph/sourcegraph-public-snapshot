package com.hello;

import com.hello1.TextContainer;
import java.util.ArrayList;

public class Text implements TextContainer {
    public Text(String chars) {
        super(chars);
    }

    public Text append(String chars, int num) {
        return new Text(this.chars + chars);
    }

    public record Person (String name, String address) {}

    enum Flags {
        Flags() {
            this(1);
        }

        Flags(int bits) {
            this.bits = bits;
        }
    }

    protected String toStringAttributes() {
        return "text=" + getChars();
    }

    public void print(int i) {
        for (int i = 0; i < 5; i++) {
          System.out.println(i);
        }
        System.out.println(i);
    }

    public interface Hello {

        public void func1(Hello t) {
            var newT = t;
        }

    }

    public void test() {
        ArrayList<Integer> numbers = new ArrayList<Integer>();
        numbers.add(5);
        numbers.add(9);
        numbers.add(8);
        numbers.add(1);
        numbers.forEach( (n) -> { System.out.println(n); } );

    }
}

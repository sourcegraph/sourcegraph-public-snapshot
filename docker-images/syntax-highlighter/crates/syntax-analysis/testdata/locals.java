package com.hello;

import java.lang.AutoCloseable;
import java.util.*;
import java.util.stream.*;

@Deprecated
public class Locals<Container> implements AutoCloseable {

	final String chars;

	public Locals(String chars) {
		this.chars = chars;
	}

	public Locals append(String chars, int num, Locals text) {
		return new Locals(this.chars + chars + text.getChars());
	}

	public String getChars() {
		return this.chars;
	}

	public void close() {
	}

	public static void create() {
		var x = new Locals<Integer>("hello");
	}

	public record Person(String name, String address) {
	}

	private class Binary<N extends Number> {
		final N val;

		public Binary(N value) {
			this.val = value;
		}
	}

	public void checks(Object person) {
		if (person instanceof Person(String x, String y)) {
			System.out.println(x + "," + y);
    }
	}

	enum Flags {
		NODE_TEXT, FOR_HEADING_ID, NO_TRIM_REF_TEXT_START, NO_TRIM_REF_TEXT_END, ADD_SPACES_BETWEEN_NODES,;

		final int bits;

		Flags() {
			this(1);
		}

		Flags(int bits) {
			this.bits = bits;
		}

		public static boolean hasNodeText(Flags bits) {
			return (bits.bits & Flags.NODE_TEXT.bits) != 0;
		}
	}

	protected String toStringAttributes() {
		return "text=" + getChars();
	}

	public <T extends Container> List<T> fromArrayToList(T[] a) {
		return Arrays.stream(a).collect(Collectors.toList());
	}

	// ? in (wildcard) node doesn't have its own node and
	// is not treated as a type identifier
	public void printList(List<? extends Container> a) {
		System.out.println(a);
	}

	public void print(int r) {
		for (int i = 0; i < r; i++) {
			System.out.println(i);
		}
		System.out.println(r);
	}

	public interface Hello {
		public void func1(Hello t);
	}

	public class Hello2 {
		public Hello2(int t) {
			var newT = t;
		}
	}

	public void blocks(int num) {
		{
			var num2 = 25;
			{
				var num3 = 100;
			}
		}
	}

	public void test() {
		ArrayList<Integer> numbers = new ArrayList<Integer>();
		numbers.add(5);
		numbers.add(9);
		numbers.add(8);
		numbers.add(1);
		numbers.forEach((n) -> {
			System.out.println(n);
		});

		for (Integer num : numbers) {
			System.out.println(num);
		}

	}
}

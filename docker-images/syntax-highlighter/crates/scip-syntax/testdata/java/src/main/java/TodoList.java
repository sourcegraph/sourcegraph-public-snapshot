package com.sourcegraph.graph.twodwo;

import java.util.ArrayList;

public class TodoList {

	public static class Item {
		private String title;
		private boolean done;

		public Item(String title) {
			this.title = title;
			this.done = false;
		}

		public Item(String title, boolean done) {
			this.title = title;
			this.done = done;
		}

		public String getTitle() {
			return this.title;
		}
	}

	private ArrayList<Item> todos;

	public TodoList() {
		todos = new ArrayList<>();
	}

	public void addTodo(Item todo) {
		todos.add(todo);
	}

	public ArrayList<Item> getTodos() {
		return todos;
	}

	public void summarise() {
    System.out.println("TODO:");
		for (Item item : todos) {
			if (item.done)
				System.out.println(" [x] " + item.getTitle());
			else
				System.out.println(" [_] " + item.getTitle());
		}
	}
}

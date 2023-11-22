package com.sourcegraph.graph.twodwo;


public class Main {

    public static void main(String[] args) {

        var todoList = new TodoList();

        for(String i: args) {
            todoList.addTodo(new TodoList.Item(i));
        }

        todoList.addTodo(new TodoList.Item("balling", true));
        todoList.addTodo(new TodoList.Item("winning"));


        todoList.summarise();

    }

}

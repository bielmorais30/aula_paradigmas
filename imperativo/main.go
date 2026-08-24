package main

import "fmt"

type Task struct {
	Id int
	Title string
	Completed bool
}

func main(){

	

	var title string
	var aux string
	var completed bool
	
	tasks := []Task{}

	// tarefas := []Task{}
	fmt.Print("Digite o titulo da tarefa:\n")
	fmt.Scan(&title)
	fmt.Print("\nEsta completada?: Sim ou Não  \n")
	fmt.Scan(&aux)

	if aux == "Sim"{
		completed = true
	}else{
		completed = false
	}

	tasks = append(tasks, Task{
		Id: 1,
		Title: title,
		Completed: completed,
	})

	fmt.Println(tasks[0].Title, tasks[0].Completed)


}

package main

import "fmt"

type Task struct {
	Id int
	Title string
	Completed bool
}

func main(){

	tasks := []Task{}
	for {
		

		var loop string

		var title string
		var aux string
		var completed bool
		

		// tarefas := []Task{}
		fmt.Print("\nDeseja adicionar uma tarefa? Sim ? ")
		fmt.Scan(&loop)

		if loop != "Sim"{
			break
		}

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

		for _, task := range tasks {
			
			fmt.Println(task.Title, task.Completed)
		}



	}

}

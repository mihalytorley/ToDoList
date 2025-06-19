You can empty the task_list.csv to start a new to_do_list

To add an item to the to_do_list, in the terminal input: go run main.go --name name_of_task --status status_of_task
-- example: "go run main.go --name task_1 --status not_started"

To change the status of a task simply write the name of the task whose status you want to change in for name and the status that that you want to change it to
-- example: "go run main.go --name task_1 --status read_to_start"

To delete a task write the name of the task that you want to delete and "delete" for the status of the task
-- example: "go run main.go --name task_1 --status delete"

To run tests run: "go test"

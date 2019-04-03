package mongo

import (
	"fmt"
	"testing"
)

func TestNewPriorityQueue(t *testing.T) {
	queue := NewPriorityQueue("mongodb://127.0.0.1:27017", "myqueue", "w1")
	queue.addTask("1", "do 1", 1)
	queue.addTask("2", "do 1",1)
	queue.addTask("3", "do 1",2)
	queue.setTaskPriority("1", 5)
	rs:=queue.getTask()

	fmt.Println(rs.Lookup("id").String())

	queue.endTask("2",true,"go")
}
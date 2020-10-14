package service

import (
	"math/rand"
	"sync"

	"github.com/pkg/errors"

	"iris/pkg/define"
)

//  singleton service
var (
	once sync.Once
	ser  *DemoService
)

type DemoService struct {
}

func GetSingleTon() *DemoService {
	once.Do(
		func() {
			ser = &DemoService{}
		})

	return ser
}

func (d *DemoService) Demook() string {
	return "hello iris"
}

func (d *DemoService) Demook2() (string, error) {

	if rand.Intn(10) == 5 {
		return "", define.ErrInnerServer("get a number", errors.Errorf("I don't like 5"))
	}
	return "hello iris", nil
}

func (d *DemoService) Demook3() (string, error) {

	if rand.Intn(10) != 8 {
		return "", define.ErrInnerServer("get a number", errors.Errorf("I Want 8"))

	}
	return "hello iris", nil
}

/////////////
// new service
type ToDoService interface {
	Get()
}

type toDoService struct {
}

func NewToDoService() ToDoService {
	return &toDoService{}

}

func (c *toDoService) Get() {
	return
}

package service

import (
	"sync"

	"goweb/iriscore/resource"
)

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
	return resource.SigleTon().YourFunction("hello gu long")
	//return "hello iris"
}

func (d *DemoService) Demook2() string {
	return resource.SigleTon().YourFunction("hello jin yong")
	//return "hello iris"
}

package main

import "net/http"

type Service interface {
	DoSomethingV1()
}

type MyService struct {
}

func (d *MyService) DoSomethingV1() {
	//TODO implement me
	panic("implement me")
}

var DefaultService Service = &MyService{}

func DoSomethingV1() {
	DefaultService.DoSomethingV1()
}

type A struct {
	b B
}

func NewA() *A {
	return &A{
		b: &B1{
			client: http.DefaultClient,
		},
	}
}

type B interface {
	DoSomethingV1()
}

type B1 struct {
	client *http.Client
}

func (b *B1) DoSomethingV1() {
	//TODO implement me
	panic("implement me")
}

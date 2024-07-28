package main

import (
	"fmt"
	"time"
)

func NewUser() {
	u := User{}
	println(u.Name)
	fmt.Printf("%v \n", u)
	fmt.Printf("%+v \n", u)
	var u1 User
	println(u1.Name)

	// up 是一个指针
	up := &User{}
	fmt.Printf("up %+v \n", up)
	up2 := new(User)
	println(up2.FirstName)
	fmt.Printf("up2 %+v \n", up2)

	u4 := User{Name: "Tom", Age: 0}
	u5 := User{"Tom", "FirstName", 0}

	u4.Name = "Jerry"
	u5.Age = 18

	var up3 *User
	// nil 上访问字段，或者方法
	println(up3.FirstName)
	println(up3)
}

type User struct {
	Name      string
	FirstName string
	Age       int
}

func (u User) ChangeName(name string) {
	fmt.Printf("change name 中 u 的地址 %p \n", &u)
	u.Name = name
}

func ChangeName(u User, name string) {

}

func (u *User) ChangeAge(age int) {
	fmt.Printf("change age 中 u 的地址 %p \n", u)
	u.Age = age
}

func ChangeAge(u *User, age int) {

}

func ChangeUser() {
	u1 := User{Name: "Tom", Age: 18}
	fmt.Printf("u1 的地址 %p \n", &u1)
	// (&u1).ChangeAge(35)
	u1.ChangeAge(35)
	// 这一步执行的时候，其实相当于复制了一个 u1，改的是复制体
	// 所以 u1 原封不动
	u1.ChangeName("Jerry")
	fmt.Printf("%+v", u1)

	up1 := &User{}
	// (*up1).ChangeName("Jerry")
	up1.ChangeName("Jerry")
	up1.ChangeAge(35)
	fmt.Printf("%+v", up1)
}

type Fish struct {
	Name string
}

func (f Fish) Swim() {
	println("fish 在游")
}

// 衍生类型可以访问原类型的字段，但是不能使用原类型的方法
// 方法是定义在上面的，内存布局一样，所以能访问字段

type FakeFish Fish

func UseFish() {
	f1 := Fish{}
	f2 := FakeFish(f1)
	//f2.Swim
	f2.Name = "Tom"
	println(f1.Name)
	var y Yu
	y.Name = "yu"
	y.Swim()
}

type MyTime time.Time

func (m MyTime) MyFunc() {

}

// 向后兼容
type Yu = Fish

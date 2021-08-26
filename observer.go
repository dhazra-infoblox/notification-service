package main

import "container/list"

// Observer design pattern is used because in Future there can be more Observers.
// e.g On reaching the Limit Notify Users as well as send log Data for Analytics or may be do some other processing.
// In that case we will create more structs like 'Topic' with an Observable object and no need to make changes in existing code.

type Observable struct {
	subs *list.List
}

func (o *Observable) Subscribe(x Observer) {
	o.subs.PushBack(x)
}

func (o *Observable) Unsubscribe(x Observer) {
	for z := o.subs.Front(); z != nil; z = z.Next() {
		if z.Value.(Observer) == x {
			o.subs.Remove(z)
		}
	}
}

func (o *Observable) Fire(data interface{}) {
	for z := o.subs.Front(); z != nil; z = z.Next() {
		z.Value.(Observer).Notify(data)
	}
}

type Observer interface {
	Notify(data interface{})
}

type LimitReached struct {
	LogType LoggerType
	Value   interface{}
}

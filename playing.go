package main

import (
	"container/list"
)

type PlayableItem struct {
	Id string
}

type Queue struct {
	videos list.List
}

func (q *Queue) GetPlayingItem() PlayableItem {
	return q.videos.Back().Value.(PlayableItem)
}

func (q *Queue) AddItem(item PlayableItem) {
	q.videos.PushBack(item)
}

func (q *Queue) String() string {
	s := ""
	for e := q.videos.Front() ; e != nil ; e = e.Next() {
		var p PlayableItem = e.Value.(PlayableItem)
		s += ";" + p.Id
	}
	return s
}
package main

import (
	"container/list"
	// "time"
)

type PlayableItem struct {
	// The YouTube id of this item.
	// TODO(kanassar): Make it such that this doesn't just represent a YouTube id.
	// Exported fields are always capitalized
	Id string `json:"id"`
	// Time that the item was started
	// start time.Time
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
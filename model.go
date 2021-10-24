package main

type Job struct {
	Filename string `json:"filename" firestore:"filename"`
}

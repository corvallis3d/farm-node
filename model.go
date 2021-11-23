package main

type Job struct {
	JobId        string
	NumberOfJobs int
	GcodeFiles   []GcodeFile `firestore:"gcode"`
	Status       int         `firestore:"status"`
}

type GcodeFile struct {
	Filename string  `firestore:"filename"`
	Time     float64 `firestore:"time"`
	Status   int     `firestore:"status"`
	Filament `firestore:"filament"`
	MaxDim   `firestore:"max_dim"`
}

type Filament struct {
	Color    string `firestore:"color"`
	Material string `firestore:"material"`
	Process  string `firestore:"process"`
}

type MaxDim struct {
	Height float64 `firestore:"height"`
	Length float64 `firestore:"length"`
	Width  float64 `firestore:"width"`
}
